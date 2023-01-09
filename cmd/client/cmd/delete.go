package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/r4start/goph-keeper/cmd/client/cfg"
	"github.com/r4start/goph-keeper/internal/client"
	"github.com/r4start/goph-keeper/internal/client/grpc"
	"github.com/r4start/goph-keeper/internal/client/storage"
)

type DeleteCommand struct {
	*cobra.Command
	config  *cfg.Config
	storage storage.Storage
}

func NewDeleteCommand(c *cfg.Config, storage storage.Storage) (*DeleteCommand, error) {
	self := &DeleteCommand{
		Command: &cobra.Command{
			Use:   "delete",
			Short: "Delete resources",
		},
		config:  c,
		storage: storage,
	}

	self.RunE = self.run
	return self, nil
}

func (s *DeleteCommand) run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	c, err := grpc.NewGrpcClient(&s.config.Server)
	if err != nil {
		return err
	}

	deleter := client.NewDeleter(c, s.storage)
	return deleter.Delete(ctx, args)
}
