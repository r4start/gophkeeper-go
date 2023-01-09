package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/r4start/goph-keeper/cmd/client/cfg"
	"github.com/r4start/goph-keeper/internal/client/grpc"
	"github.com/r4start/goph-keeper/internal/client/storage"
	"github.com/r4start/goph-keeper/internal/crypto"
)

type RegisterCommand struct {
	*cobra.Command
	config  *cfg.Config
	storage storage.Storage
}

func NewRegisterCommand(c *cfg.Config, storage storage.Storage) (*RegisterCommand, error) {
	self := &RegisterCommand{
		Command: &cobra.Command{
			Use:   "register",
			Short: "Register a new account in gophkeeper.",
			Long:  `Register a new account in gophkeeper. Provide login, password and master password to complete registration.`,
		},
		config:  c,
		storage: storage,
	}

	self.RunE = self.run

	self.Flags().StringP(CmdFlagLogin, "l", "", "User login.")
	self.Flags().StringP(CmdFlagPassword, "p", "", "User password.")
	self.Flags().StringP(CmdFlagMasterPassword, "m", "", "Master password.")

	if err := self.MarkFlagRequired(CmdFlagLogin); err != nil {
		return nil, err
	}
	if err := self.MarkFlagRequired(CmdFlagPassword); err != nil {
		return nil, err
	}
	if err := self.MarkFlagRequired(CmdFlagMasterPassword); err != nil {
		return nil, err
	}

	return self, nil
}

func (r *RegisterCommand) run(cmd *cobra.Command, args []string) error {
	l, err := cmd.Flags().GetString(CmdFlagLogin)
	if err != nil {
		return err
	}

	p, err := cmd.Flags().GetString(CmdFlagPassword)
	if err != nil {
		return err
	}

	mp, err := cmd.Flags().GetString(CmdFlagMasterPassword)
	if err != nil {
		return err
	}

	s, err := crypto.GenerateMasterKey([]byte(mp))
	if err != nil {
		return err
	}

	c, err := grpc.NewGrpcClient(&r.config.Server)
	if err != nil {
		return err
	}

	auth, err := c.Register(context.Background(), l, p, s.Salt)
	if err != nil {
		return err
	}

	ud := &storage.UserData{
		UserID:       auth.UserID,
		Token:        auth.Token,
		RefreshToken: auth.RefreshToken,
		MasterKey:    s.Key,
		Salt:         s.Salt,
	}

	if err := r.storage.SetUserData(context.Background(), ud); err != nil {
		return err
	}
	return nil
}
