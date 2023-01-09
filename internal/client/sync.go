package client

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"

	pb "github.com/r4start/goph-keeper/pkg/client/proto"

	"github.com/r4start/goph-keeper/internal/client/storage"
	"github.com/r4start/goph-keeper/internal/crypto"
)

const (
	_operationTimeout = 500 * time.Millisecond

	ResourceTypeBinary = resourceType(iota)
	ResourceTypeCredentials
	ResourceTypeCardCredentials
)

type resourceType int

type resourcePair struct {
	ID   string
	Type resourceType
}

type SynchronizerOption func(*Synchronizer)

type Synchronizer struct {
	client           Client
	storage          storage.Storage
	syncDirectory    string
	limit            int
	operationTimeout time.Duration
}

func NewSynchronizer(
	client Client,
	storage storage.Storage,
	syncDirectory string,
	opts ...SynchronizerOption,
) *Synchronizer {
	s := &Synchronizer{
		client:           client,
		storage:          storage,
		syncDirectory:    syncDirectory,
		limit:            1,
		operationTimeout: _operationTimeout,
	}

	for _, op := range opts {
		op(s)
	}

	return s
}

func (s *Synchronizer) Sync(ctx context.Context) error {
	userData, err := s.storage.UserData(ctx)
	if err != nil {
		return err
	}

	auth := &UserAuthorization{
		Token:  userData.Token,
		UserID: userData.UserID,
	}

	group, grctx := errgroup.WithContext(ctx)
	group.SetLimit(s.limit)

	serverFiles := make(map[string]bool)
	local := make(map[string]resourcePair)

	group.Go(func() error {
		rfs, err := s.getRemoteResources(grctx, auth)
		if err != nil {
			return err
		}
		for _, e := range rfs {
			serverFiles[e.ID] = true
		}
		return nil
	})
	group.Go(
		func() error {
			var err error
			local, err = listLocalResources(grctx, s.storage)
			return err
		})

	if err := group.Wait(); err != nil {
		return err
	}

	toDownload, toDelete := createSyncLists(local, serverFiles)

	errCh := make(chan error)
	go func() {
		err := s.downloadResources(ctx, auth, toDownload)
		errCh <- err
	}()

	go func() {
		err := s.deleteResources(ctx, toDelete)
		errCh <- err
	}()

	for i := 0; i < 2; i++ {
		if e := <-errCh; e != nil {
			err = multierror.Append(err, e)
		}
	}
	return err
}

func (s *Synchronizer) deleteFile(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, s.operationTimeout)
	defer cancel()
	data, err := s.storage.FileData(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to retrieve file [%s] data: %w", id, err)
	}

	err = s.storage.DeleteFile(ctx, data)
	if err != nil {
		return fmt.Errorf("failed to delete file [%s] from storage: %w", id, err)
	}

	if err := os.Remove(data.Path); err != nil {
		return fmt.Errorf("failed to delete file [%s] from %s: %w", id, data.Path, err)
	}
	return nil
}

func (s *Synchronizer) deletePassword(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, s.operationTimeout)
	defer cancel()

	if err := s.storage.DeleteCredential(ctx, id); err != nil {
		return fmt.Errorf("failed to delete creds [%s] from storage: %w", id, err)
	}

	return nil
}

func (s *Synchronizer) deleteCardCredentials(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, s.operationTimeout)
	defer cancel()

	if err := s.storage.DeleteCard(ctx, id); err != nil {
		return fmt.Errorf("failed to delete card [%s] from storage: %w", id, err)
	}

	return nil
}

func (s *Synchronizer) getRemoteResources(ctx context.Context, auth *UserAuthorization) ([]ResourceInfo, error) {
	reader, err := s.client.List(ctx, auth)
	if err != nil {
		return nil, err
	}

	result := make([]ResourceInfo, 0, 128)
	for {
		opCtx, cancel := context.WithTimeout(ctx, s.operationTimeout)
		i, err := reader.Recv(opCtx)
		cancel()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		result = append(result, *i)
	}

	return result, nil
}

func (s *Synchronizer) deleteResources(ctx context.Context, ids []resourcePair) error {
	var result error
	for _, e := range ids {
		switch e.Type {
		case ResourceTypeBinary:
			if err := s.deleteFile(ctx, e.ID); err != nil {
				result = multierror.Append(result, err)
			}
		case ResourceTypeCardCredentials:
			if err := s.deleteCardCredentials(ctx, e.ID); err != nil {
				result = multierror.Append(result, err)
			}
		case ResourceTypeCredentials:
			if err := s.deletePassword(ctx, e.ID); err != nil {
				result = multierror.Append(result, err)
			}
		}
	}

	return result
}

func (s *Synchronizer) downloadResources(ctx context.Context, auth *UserAuthorization, ids []string) error {
	ud, err := s.storage.UserData(ctx)
	if err != nil {
		return err
	}

	errCh := make(chan error)
	for _, id := range ids {
		go func(id string) {
			errCh <- s.downloadResource(ctx, auth, ud, id)
		}(id)
	}

	for i := 0; i < len(ids); i++ {
		if e := <-errCh; e != nil {
			err = multierror.Append(err, e)
		}
	}
	return err
}

func (s *Synchronizer) downloadResource(ctx context.Context, auth *UserAuthorization, ud *storage.UserData, id string) error {
	downloader, err := s.client.Get(ctx, auth, id)
	if err != nil {
		return fmt.Errorf("failed to get resource %s: %w", id, err)
	}
	defer func() {
		_ = downloader.Close()
	}()

	var (
		salt     []byte
		fileSize uint64
	)

	buffer := make([]byte, 0)
	for {
		chunk, err := downloader.Recv(ctx)
		if err != nil {
			return fmt.Errorf("failed to download resource %s: %w", id, err)
		}

		if chunk.Salt != nil {
			if salt != nil {
				return fmt.Errorf("salt has been already received")
			}
			salt = chunk.Salt
			continue
		}
		if chunk.Size != nil {
			if fileSize != *chunk.Size {
				return fmt.Errorf("bad files size: %d recvd; %d sent", fileSize, *chunk.Size)
			}
			break
		}

		fileSize += uint64(len(chunk.Data))
		buffer = append(buffer, chunk.Data...)
	}

	decoder, err := crypto.RestoreAesGcmEncoder(ud.MasterKey, salt)
	if err != nil {
		return fmt.Errorf("failed to restore decoder for resource %s: %w", id, err)
	}

	buffer, err = decoder.Decode(ctx, buffer)
	if err != nil {
		return fmt.Errorf("failed to decode resource %s: %w", id, err)
	}

	var dr pb.DataResource
	if err := proto.Unmarshal(buffer, &dr); err != nil {
		return fmt.Errorf("failed to unmarshal data resource %s: %w", id, err)
	}

	switch *dr.Type {
	case pb.DataType_DATA_TYPE_BINARY:
		destPath := s.syncDirectory + string(os.PathSeparator) + *dr.Name
		f, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("failed to create output file %s for resource %s: %w", destPath, id, err)
		}

		wr, err := f.Write(dr.Data)
		if err != nil || wr != len(dr.Data) {
			if e := os.Remove(destPath); e != nil {
				err = multierror.Append(err, e)
			}
			return fmt.Errorf("failed to write into a file %s: resource %s %w", destPath, id, err)
		}

		if err := f.Close(); err != nil {
			return fmt.Errorf("failed to close a file %s: resource %s %w", destPath, id, err)
		}

		err = s.storage.AddFile(ctx, &storage.FileData{
			ID:     id,
			UserID: auth.UserID,
			Name:   *dr.Name,
			Path:   destPath,
			Key: crypto.Secret{
				Key:  decoder.Key(),
				Salt: decoder.Salt(),
			},
		})
		if err != nil {
			return fmt.Errorf("failed to save file info for a resource %s %w", id, err)
		}
	case pb.DataType_DATA_TYPE_CARD_CREDENTIALS:
		var cardData pb.CardData
		err := proto.Unmarshal(dr.Data, &cardData)
		if err != nil {
			return fmt.Errorf("failed to unmarshal card credentials. resource: %s %w", id, err)
		}

		name := ""
		if dr.Name != nil {
			name = *dr.Name
		}

		err = s.storage.AddCard(ctx, &storage.CardData{
			ID:           id,
			UserID:       auth.UserID,
			Name:         name,
			Number:       *cardData.Number,
			Holder:       *cardData.Cardholder,
			ExpiryDate:   *cardData.ExpiryDate,
			SecurityCode: *cardData.SecurityCode,
		})
		if err != nil {
			return fmt.Errorf("failed to save card credentials. resource: %s %w", id, err)
		}
	case pb.DataType_DATA_TYPE_CREDENTIALS:
		var pwdData pb.PasswordData
		err := proto.Unmarshal(dr.Data, &pwdData)
		if err != nil {
			return fmt.Errorf("failed to unmarshal credentials. resource: %s %w", id, err)
		}

		desc := ""
		if pwdData.Description != nil {
			desc = *pwdData.Description
		}

		err = s.storage.AddCredentials(ctx, &storage.CredentialData{
			ID:          id,
			UserID:      auth.UserID,
			Username:    *pwdData.Username,
			Password:    *pwdData.Password,
			Uri:         *pwdData.Uri,
			Description: desc,
		})
		if err != nil {
			return fmt.Errorf("failed to save credentials. resource: %s %w", id, err)
		}
	}
	return nil
}

func listLocalResources(ctx context.Context, storage storage.Storage) (map[string]resourcePair, error) {
	var (
		localResources = make(chan resourcePair)
		errCh          = make(chan error)
		exit           = make(chan any)
		wg             sync.WaitGroup
		err            error
	)

	wg.Add(3)

	go func() {
		defer wg.Done()
		fs, err := storage.ListFiles(ctx)
		if err != nil {
			errCh <- err
			return
		}
		for _, e := range fs {
			localResources <- resourcePair{
				ID:   e.ID,
				Type: ResourceTypeBinary,
			}
		}
	}()

	go func() {
		defer wg.Done()
		cards, err := storage.ListCards(ctx)
		if err != nil {
			errCh <- err
		}
		for _, e := range cards {
			localResources <- resourcePair{
				ID:   e.ID,
				Type: ResourceTypeCardCredentials,
			}
		}
	}()

	go func() {
		defer wg.Done()
		creds, err := storage.ListCredentials(ctx)
		if err != nil {
			errCh <- err
			return
		}
		for _, e := range creds {
			localResources <- resourcePair{
				ID:   e.ID,
				Type: ResourceTypeCredentials,
			}
		}
	}()

	go func() {
		wg.Wait()
		close(exit)
	}()

	local := make(map[string]resourcePair)

outerloop:
	for {
		select {
		case res := <-localResources:
			local[res.ID] = res
		case e := <-errCh:
			err = multierror.Append(err, e)
		case <-exit:
			break outerloop
		}
	}
	return local, nil
}

func createSyncLists(
	localResources map[string]resourcePair,
	remoteResources map[string]bool,
) (
	[]string, []resourcePair,
) {
	remoteLen := len(remoteResources)
	localLen := len(localResources)
	toDelete := make([]resourcePair, 0, max(remoteLen, localLen))
	toDownload := make([]string, 0, max(remoteLen, localLen))

	for e := range remoteResources {
		if _, ok := localResources[e]; !ok {
			toDownload = append(toDownload, e)
		}
		delete(localResources, e)
	}

	for _, v := range localResources {
		toDelete = append(toDelete, v)
	}
	return toDownload, toDelete
}

func WithConcurrencyLimit(limit int) SynchronizerOption {
	return func(s *Synchronizer) {
		s.limit = limit
	}
}

func WithOperationTimeout(t time.Duration) SynchronizerOption {
	return func(s *Synchronizer) {
		s.operationTimeout = t
	}
}

type number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

func max[T number](lhs, rhs T) T {
	if lhs < rhs {
		return rhs
	}
	return lhs
}
