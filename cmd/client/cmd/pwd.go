package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/r4start/goph-keeper/cmd/client/cfg"
	"github.com/r4start/goph-keeper/internal/client"
	"github.com/r4start/goph-keeper/internal/client/grpc"
	"github.com/r4start/goph-keeper/internal/client/storage"
)

const (
	_pwdUsername = "username"
	_pwdPassword = "password"
	_pwdURI      = "uri"
	_pwdInfo     = "info"
)

type PwdCommand struct {
	*cobra.Command
	config  *cfg.Config
	storage storage.Storage
}

func NewPwdCommand(c *cfg.Config, storage storage.Storage) (*PwdCommand, error) {
	self := &PwdCommand{
		Command: &cobra.Command{
			Use:   "cred",
			Short: "Store creadential data securely in Gophkeeper.",
		},
		config:  c,
		storage: storage,
	}

	self.RunE = self.run

	self.Flags().StringP(_pwdUsername, "l", "", "Login.")
	self.Flags().StringP(_pwdPassword, "p", "", "Password.")
	self.Flags().StringP(_pwdURI, "u", "", "Holder's name.")
	self.Flags().StringP(_pwdInfo, "d", "", "Description.")

	if err := self.MarkFlagRequired(_pwdUsername); err != nil {
		return nil, err
	}

	if err := self.MarkFlagRequired(_pwdPassword); err != nil {
		return nil, err
	}

	if err := self.MarkFlagRequired(_pwdURI); err != nil {
		return nil, err
	}

	return self, nil
}

func (s *PwdCommand) run(cmd *cobra.Command, args []string) error {
	username, err := s.Flags().GetString(_pwdUsername)
	if err != nil {
		return err
	}

	pwd, err := s.Flags().GetString(_pwdPassword)
	if err != nil {
		return err
	}

	uri, err := s.Flags().GetString(_pwdURI)
	if err != nil {
		return err
	}

	info, err := s.Flags().GetString(_pwdInfo)
	if err != nil {
		return err
	}

	c, err := grpc.NewGrpcClient(&s.config.Server)
	if err != nil {
		return err
	}

	uploader := client.NewUploader(c, s.storage, s.config.SyncDirectory)
	err = uploader.UploadCredentials(context.Background(), storage.CredentialData{
		Username:    username,
		Password:    pwd,
		Uri:         uri,
		Description: info,
	})

	return err
}
