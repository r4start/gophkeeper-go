package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/r4start/goph-keeper/cmd/client/cfg"
	"github.com/r4start/goph-keeper/internal/client"
	"github.com/r4start/goph-keeper/internal/client/grpc"
	"github.com/r4start/goph-keeper/internal/client/storage"
)

type SyncCommand struct {
	*cobra.Command
	config  *cfg.Config
	storage storage.Storage
}

func NewSyncCommand(c *cfg.Config, storage storage.Storage) (*SyncCommand, error) {
	self := &SyncCommand{
		Command: &cobra.Command{
			Use:   "sync",
			Short: "Sync files with a server.",
		},
		config:  c,
		storage: storage,
	}

	self.RunE = self.run
	return self, nil
}

func (s *SyncCommand) run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	c, err := grpc.NewGrpcClient(&s.config.Server)
	if err != nil {
		return err
	}
	syncer := client.NewSynchronizer(c, s.storage, s.config.SyncDirectory)
	return syncer.Sync(ctx)
}
