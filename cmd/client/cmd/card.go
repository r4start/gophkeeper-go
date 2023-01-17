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
	_cardName         = "name"
	_cardNumber       = "number"
	_cardHolder       = "holder"
	_cardSecurityCode = "seccode"
	_cardExpiryDate   = "expdate"
)

type CardCommand struct {
	*cobra.Command
	config  *cfg.Config
	storage storage.Storage
}

func NewCardCommand(c *cfg.Config, storage storage.Storage) (*CardCommand, error) {
	self := &CardCommand{
		Command: &cobra.Command{
			Use:   "card",
			Short: "Store card data securely in Gophkeeper.",
		},
		config:  c,
		storage: storage,
	}

	self.RunE = self.run

	self.Flags().StringP(_cardName, "i", "", "Card identificator/name.")
	self.Flags().StringP(_cardNumber, "n", "", "Card number.")
	self.Flags().StringP(_cardHolder, "u", "", "Holder's name.")
	self.Flags().StringP(_cardExpiryDate, "d", "", "Expiry date in MM/YY format")
	self.Flags().StringP(_cardSecurityCode, "c", "", "Security code if present.")

	if err := self.MarkFlagRequired(_cardName); err != nil {
		return nil, err
	}

	if err := self.MarkFlagRequired(_cardNumber); err != nil {
		return nil, err
	}

	if err := self.MarkFlagRequired(_cardHolder); err != nil {
		return nil, err
	}

	if err := self.MarkFlagRequired(_cardExpiryDate); err != nil {
		return nil, err
	}

	return self, nil
}

func (s *CardCommand) run(cmd *cobra.Command, args []string) error {
	name, err := s.Flags().GetString(_cardName)
	if err != nil {
		return err
	}

	number, err := s.Flags().GetString(_cardNumber)
	if err != nil {
		return err
	}

	holder, err := s.Flags().GetString(_cardHolder)
	if err != nil {
		return err
	}

	expDate, err := s.Flags().GetString(_cardExpiryDate)
	if err != nil {
		return err
	}

	cvc, err := s.Flags().GetString(_cardSecurityCode)
	if err != nil {
		return err
	}

	c, err := grpc.NewGrpcClient(&s.config.Server)
	if err != nil {
		return err
	}

	uploader := client.NewUploader(c, s.storage, s.config.SyncDirectory)
	err = uploader.UploadCard(context.Background(), storage.CardData{
		Name:         name,
		Number:       number,
		Holder:       holder,
		ExpiryDate:   expDate,
		SecurityCode: cvc,
	})
	return err
}
