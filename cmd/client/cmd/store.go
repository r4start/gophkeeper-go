package cmd

import (
	"github.com/spf13/cobra"

	"github.com/r4start/goph-keeper/cmd/client/cfg"

	"github.com/r4start/goph-keeper/internal/client/storage"
)

type StoreCommand struct {
	*cobra.Command
	config  *cfg.Config
	storage storage.Storage
}

func NewStoreCommand(c *cfg.Config, storage storage.Storage) (*StoreCommand, error) {
	self := &StoreCommand{
		Command: &cobra.Command{
			Use:   "store",
			Short: "Securely store data in Gophkeeper.",
		},
		config:  c,
		storage: storage,
	}

	filesCmd, err := NewStoreFilesCommand(c, storage)
	if err != nil {
		return nil, err
	}

	cardsCmd, err := NewCardCommand(c, storage)
	if err != nil {
		return nil, err
	}

	pwdCmd, err := NewPwdCommand(c, storage)
	if err != nil {
		return nil, err
	}

	self.AddCommand(filesCmd.Command)
	self.AddCommand(cardsCmd.Command)
	self.AddCommand(pwdCmd.Command)

	return self, nil
}
