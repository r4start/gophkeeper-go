package client

import (
	"context"
	"os"

	"github.com/r4start/goph-keeper/internal/client/storage"
)

type Deleter struct {
	client  Client
	storage storage.Storage
}

func NewDeleter(client Client, storage storage.Storage) *Deleter {
	return &Deleter{
		client:  client,
		storage: storage,
	}
}

func (d *Deleter) Delete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	userData, err := d.storage.UserData(ctx)
	if err != nil {
		return err
	}

	auth := &UserAuthorization{
		Token:  userData.Token,
		UserID: userData.UserID,
	}

	localResources, err := listLocalResources(ctx, d.storage)

	toDelete := make([]resourcePair, 0, len(ids))
	for _, arg := range ids {
		if _, ok := localResources[arg]; ok {
			toDelete = append(toDelete, localResources[arg])
		}
	}

	for _, res := range toDelete {
		if err := d.client.Delete(ctx, auth, res.ID); err != nil {
			return err
		}

		switch res.Type {
		case ResourceTypeBinary:
			data, err := d.storage.FileData(ctx, res.ID)
			if err != nil {
				return err
			}

			if err := d.storage.DeleteFile(ctx, data); err != nil {
				return err
			}
			if err := os.Remove(data.Path); err != nil {
				return err
			}
		case ResourceTypeCredentials:
			if err := d.storage.DeleteCredential(ctx, res.ID); err != nil {
				return err
			}
		case ResourceTypeCardCredentials:
			if err := d.storage.DeleteCard(ctx, res.ID); err != nil {
				return err
			}
		}
	}

	return nil
}
