package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"

	"github.com/hashicorp/go-multierror"
	pb "github.com/r4start/goph-keeper/pkg/client/proto"

	"github.com/r4start/goph-keeper/internal/client/storage"
	"github.com/r4start/goph-keeper/internal/crypto"
)

const (
	_minimalFileSize = 1

	_bufferReadSize = 4 * 1024 * 1024 // 4 MiB
)

type UploaderOption func(u *Uploader)

type Uploader struct {
	client           Client
	storage          storage.Storage
	syncDirectory    string
	limit            int
	operationTimeout time.Duration
}

func NewUploader(
	client Client,
	storage storage.Storage,
	syncDirectory string,
	opts ...UploaderOption,
) *Uploader {
	u := &Uploader{
		client:           client,
		storage:          storage,
		syncDirectory:    syncDirectory,
		limit:            1,
		operationTimeout: time.Second,
	}

	for _, o := range opts {
		o(u)
	}

	return u
}

func (u *Uploader) UploadFiles(ctx context.Context, files []string) error {
	uniqFiles, err := u.filterFiles(ctx, files)
	if err != nil {
		return err
	}

	userData, err := u.storage.UserData(ctx)
	if err != nil {
		return err
	}

	userAuth := &UserAuthorization{
		Token:  userData.Token,
		UserID: userData.UserID,
	}

	gr, grctx := errgroup.WithContext(ctx)
	gr.SetLimit(u.limit)

	for _, f := range uniqFiles {
		file := f
		gr.Go(func() error {
			data, name, err := copyFile(file, u.syncDirectory)
			if err != nil {
				return err
			}

			blob, err := u.serializeFile(name, data)
			if err != nil {
				path := fmt.Sprintf("%s%c%s", u.syncDirectory, os.PathSeparator, name)
				if e := os.Remove(path); e != nil {
					err = multierror.Append(err, e)
				}
				return err
			}

			info, key, err := u.upload(grctx, userData, userAuth, blob)
			if err != nil {
				path := fmt.Sprintf("%s%c%s", u.syncDirectory, os.PathSeparator, name)
				if e := os.Remove(path); e != nil {
					err = multierror.Append(err, e)
				}
				return err
			}

			return u.saveFileInfo(grctx, info.ID, userAuth.UserID, name, *key)
		})
	}

	return gr.Wait()
}

func (u *Uploader) UploadCard(ctx context.Context, card storage.CardData) error {
	userData, err := u.storage.UserData(ctx)
	if err != nil {
		return err
	}

	userAuth := &UserAuthorization{
		Token:  userData.Token,
		UserID: userData.UserID,
	}

	card.UserID = userData.UserID
	data, err := u.serializeCard(card)
	if err != nil {
		return err
	}

	info, _, err := u.upload(ctx, userData, userAuth, data)
	if err != nil {
		return err
	}

	card.ID = info.ID
	return u.storage.AddCard(ctx, &card)
}

func (u *Uploader) UploadCredentials(ctx context.Context, cred storage.CredentialData) error {
	userData, err := u.storage.UserData(ctx)
	if err != nil {
		return err
	}

	userAuth := &UserAuthorization{
		Token:  userData.Token,
		UserID: userData.UserID,
	}

	cred.UserID = userData.UserID
	data, err := u.serializeCredentials(cred)
	if err != nil {
		return err
	}

	info, _, err := u.upload(ctx, userData, userAuth, data)
	if err != nil {
		return err
	}
	cred.ID = info.ID

	return u.storage.AddCredentials(ctx, &cred)
}

func (u *Uploader) serializeFile(name string, rawData []byte) ([]byte, error) {
	dt := pb.DataType_DATA_TYPE_BINARY
	data := &pb.DataResource{
		Type: &dt,
		Data: rawData,
		Name: &name,
	}
	return proto.Marshal(data)
}

func (u *Uploader) saveFileInfo(ctx context.Context, id string, userID string, name string, key crypto.Secret) error {
	filePath := fmt.Sprintf("%s%c%s", u.syncDirectory, os.PathSeparator, name)
	fd := &storage.FileData{
		ID:     id,
		UserID: userID,
		Name:   name,
		Path:   filePath,
		Key:    key,
	}

	return u.storage.AddFile(ctx, fd)
}

func (u *Uploader) serializeCard(card storage.CardData) ([]byte, error) {
	m := &pb.CardData{
		Number:     &card.Number,
		Cardholder: &card.Holder,
		ExpiryDate: &card.ExpiryDate,
	}
	if len(card.SecurityCode) != 0 {
		m.SecurityCode = &card.SecurityCode
	}

	rawData, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}

	dt := pb.DataType_DATA_TYPE_CARD_CREDENTIALS
	data := &pb.DataResource{
		Type: &dt,
		Data: rawData,
		Name: &card.Name,
	}

	msg, err := proto.Marshal(data)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (u *Uploader) serializeCredentials(cred storage.CredentialData) ([]byte, error) {
	m := &pb.PasswordData{
		Username:    &cred.Username,
		Password:    &cred.Password,
		Uri:         &cred.Uri,
		Description: &cred.Description,
	}

	rawData, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}

	dt := pb.DataType_DATA_TYPE_CREDENTIALS
	data := &pb.DataResource{
		Type: &dt,
		Data: rawData,
	}

	msg, err := proto.Marshal(data)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (u *Uploader) upload(
	ctx context.Context,
	user *storage.UserData,
	authData *UserAuthorization,
	data []byte,
) (
	*ResourceInfo,
	*crypto.Secret,
	error,
) {
	encoder, err := crypto.NewAesGcmEncoder(user.MasterKey)
	if err != nil {
		return nil, nil, err
	}
	msg, err := encoder.Encode(ctx, data)
	if err != nil {
		return nil, nil, err
	}

	writer, err := u.client.Store(ctx, authData, encoder.Salt(), uint64(len(msg)))
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		_ = writer.Close()
	}()

	br := bytes.NewReader(msg)
	buffer := make([]byte, 0, _bufferReadSize)
	for {
		read, err := br.Read(buffer[:cap(buffer)])
		if read == 0 {
			if err == nil {
				continue
			}
			if err == io.EOF {
				break
			}
			return nil, nil, err
		}

		if err != nil {
			return nil, nil, err
		}

		if err := writer.SendChunk(ctx, buffer[:read]); err != nil {
			return nil, nil, err
		}

	}

	result, err := writer.Recv(ctx)
	if err != nil {
		return nil, nil, err
	}

	if result.ErrorCode != 0 {
		return nil, nil, fmt.Errorf("got an error from a server: %d", result.ErrorCode)
	}
	return result, &crypto.Secret{
		Key:  encoder.Key(),
		Salt: encoder.Salt(),
	}, nil
}

func (u *Uploader) filterFiles(ctx context.Context, files []string) ([]string, error) {
	uniqFiles := make(map[string]string, len(files))
	for _, f := range files {
		name, err := fileName(f)
		if err != nil {
			return nil, err
		}
		abs, err := filepath.Abs(f)
		if err != nil {
			return nil, err
		}
		uniqFiles[name] = abs
	}

	fs, err := u.storage.ListFiles(ctx)
	if err != nil {
		return nil, err
	}

	for _, f := range fs {
		if _, ok := uniqFiles[f.Name]; ok {
			delete(uniqFiles, f.Name)
		}
	}

	result := make([]string, 0, len(uniqFiles))
	for _, v := range uniqFiles {
		result = append(result, v)
	}
	return result, nil
}

func copyFile(filePath, syncDirPath string) ([]byte, string, error) {
	name, err := fileName(filePath)
	if err != nil {
		return nil, "", err
	}

	buffer, err := os.ReadFile(filePath)
	if err != nil {
		return nil, "", err
	}

	if len(buffer) < _minimalFileSize {
		return nil, "", fmt.Errorf("bad file size: file(%s) size(%d)", filePath, len(buffer))
	}

	copyFilePath := fmt.Sprintf("%s%c%s", syncDirPath, os.PathSeparator, name)
	f, err := os.OpenFile(copyFilePath, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return nil, "", err
	}
	defer func() {
		_ = f.Close()
	}()

	if _, err := f.Write(buffer); err != nil {
		return nil, "", err
	}

	return buffer, name, nil
}

func fileName(filePath string) (string, error) {
	pathParts := strings.Split(filePath, string(os.PathSeparator))
	if len(pathParts) == 0 {
		return "", fmt.Errorf("failed to prepare file name for a path: %s", filePath)
	}
	return pathParts[len(pathParts)-1], nil
}

func WithUploaderLimit(limit int) UploaderOption {
	return func(u *Uploader) {
		u.limit = limit
	}
}

func WithUploaderOperationTimeout(t time.Duration) UploaderOption {
	return func(u *Uploader) {
		u.operationTimeout = t
	}
}
