package cmd

import (
	"github.com/spf13/cobra"
)

const (
	CmdFlagLogin          = "login"
	CmdFlagPassword       = "password"
	CmdFlagMasterPassword = "master-password"
)

type RootCommand struct {
	*cobra.Command
}

func NewRootCommand() *RootCommand {
	self := &RootCommand{
		Command: &cobra.Command{
			Use:   "gkcli",
			Short: "Gophkeeper client application",
		}}

	return self
}

func (r *RootCommand) Execute() error {
	return r.Command.Execute()
}

func (r *RootCommand) Cmd() *cobra.Command {
	return r.Command
}
