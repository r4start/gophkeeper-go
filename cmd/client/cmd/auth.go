package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/r4start/goph-keeper/cmd/client/cfg"
	"github.com/r4start/goph-keeper/internal/client"
	"github.com/r4start/goph-keeper/internal/client/grpc"
	"github.com/r4start/goph-keeper/internal/client/storage"
	"github.com/r4start/goph-keeper/internal/crypto"
)

type AuthCommand struct {
	*cobra.Command
	config  *cfg.Config
	storage storage.Storage
}

func NewAuthCommand(c *cfg.Config, storage storage.Storage) (*AuthCommand, error) {
	self := &AuthCommand{
		Command: &cobra.Command{
			Use:   "auth",
			Short: "Sign in gophkeeper account.",
			Long:  `Sign in gophkeeper account. Provide login, password and master password.`,
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

func (a *AuthCommand) run(cmd *cobra.Command, args []string) error {
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

	var c client.Client
	c, err = grpc.NewGrpcClient(&a.config.Server)
	if err != nil {
		return err
	}

	auth, err := c.Authorize(context.Background(), l, p)
	if err != nil {
		return err
	}

	key, err := crypto.RecoverMasterKey([]byte(mp), auth.Salt)
	if err != nil {
		return err
	}

	ud := &storage.UserData{
		UserID:       auth.UserID,
		Token:        auth.Token,
		RefreshToken: auth.RefreshToken,
		MasterKey:    key.Key,
		Salt:         auth.Salt,
	}

	if err := a.storage.SetUserData(context.Background(), ud); err != nil {
		panic(err)
	}
	return nil
}
