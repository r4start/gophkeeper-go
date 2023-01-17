package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/r4start/goph-keeper/cmd/client/cfg"

	"github.com/r4start/goph-keeper/internal/client"
	"github.com/r4start/goph-keeper/internal/client/grpc"
	"github.com/r4start/goph-keeper/internal/client/storage"
)

type StoreFilesCommand struct {
	*cobra.Command
	config  *cfg.Config
	storage storage.Storage
}

func NewStoreFilesCommand(c *cfg.Config, storage storage.Storage) (*StoreFilesCommand, error) {
	self := &StoreFilesCommand{
		Command: &cobra.Command{
			Use:   "files",
			Short: "Send a file to Gophkeeper.",
		},
		config:  c,
		storage: storage,
	}

	self.RunE = self.run
	return self, nil
}

func (s *StoreFilesCommand) run(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return nil
	}
	c, err := grpc.NewGrpcClient(&s.config.Server)
	if err != nil {
		return err
	}

	uploader := client.NewUploader(c, s.storage, s.config.SyncDirectory)
	return uploader.UploadFiles(context.Background(), args)

}
