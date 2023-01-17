package cmd

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	"github.com/alexeyco/simpletable"

	"github.com/r4start/goph-keeper/cmd/client/cfg"
	"github.com/r4start/goph-keeper/internal/client/storage"
)

type ListCommand struct {
	*cobra.Command
	config  *cfg.Config
	storage storage.Storage
}

func NewListCommand(c *cfg.Config, storage storage.Storage) (*ListCommand, error) {
	self := &ListCommand{
		Command: &cobra.Command{
			Use:   "list",
			Short: "List available resources.",
		},
		config:  c,
		storage: storage,
	}

	self.RunE = self.run
	return self, nil
}

func (s *ListCommand) run(cmd *cobra.Command, args []string) error {
	var (
		ctx            = context.Background()
		localResources = make(chan *simpletable.Table)
		errCh          = make(chan error)
		exit           = make(chan any)
		wg             sync.WaitGroup
		err            error
	)

	wg.Add(3)

	go func() {
		defer wg.Done()
		fs, err := s.storage.ListFiles(ctx)
		if err != nil {
			errCh <- err
			return
		}

		if len(fs) == 0 {
			return
		}

		table := simpletable.New()
		table.Header = &simpletable.Header{
			Cells: []*simpletable.Cell{
				{Align: simpletable.AlignCenter, Text: "#"},
				{Align: simpletable.AlignCenter, Text: "ID"},
				{Align: simpletable.AlignCenter, Text: "NAME"},
				{Align: simpletable.AlignCenter, Text: "PATH"},
			},
		}
		for i, e := range fs {
			row := []*simpletable.Cell{
				{Align: simpletable.AlignLeft, Text: fmt.Sprintf("%d", i+1)},
				{Align: simpletable.AlignLeft, Text: e.ID},
				{Align: simpletable.AlignLeft, Text: e.Name},
				{Align: simpletable.AlignLeft, Text: e.Path},
			}
			table.Body.Cells = append(table.Body.Cells, row)
		}

		localResources <- table
	}()

	go func() {
		defer wg.Done()
		cards, err := s.storage.ListCards(ctx)
		if err != nil {
			errCh <- err
			return
		}

		if len(cards) == 0 {
			return
		}

		table := simpletable.New()
		table.Header = &simpletable.Header{
			Cells: []*simpletable.Cell{
				{Align: simpletable.AlignCenter, Text: "#"},
				{Align: simpletable.AlignCenter, Text: "ID"},
				{Align: simpletable.AlignCenter, Text: "NUMBER"},
				{Align: simpletable.AlignCenter, Text: "HOLDER"},
				{Align: simpletable.AlignCenter, Text: "EXPIRY DATE"},
				{Align: simpletable.AlignCenter, Text: "SECURITY CODE"},
				{Align: simpletable.AlignCenter, Text: "NAME"},
			},
		}
		for i, e := range cards {
			row := []*simpletable.Cell{
				{Align: simpletable.AlignLeft, Text: fmt.Sprintf("%d", i+1)},
				{Align: simpletable.AlignLeft, Text: e.ID},
				{Align: simpletable.AlignLeft, Text: e.Number},
				{Align: simpletable.AlignLeft, Text: e.Holder},
				{Align: simpletable.AlignLeft, Text: e.ExpiryDate},
				{Align: simpletable.AlignLeft, Text: e.SecurityCode},
				{Align: simpletable.AlignLeft, Text: e.Name},
			}
			table.Body.Cells = append(table.Body.Cells, row)
		}
		localResources <- table
	}()

	go func() {
		defer wg.Done()
		creds, err := s.storage.ListCredentials(ctx)
		if err != nil {
			errCh <- err
			return
		}

		if len(creds) == 0 {
			return
		}

		table := simpletable.New()
		table.Header = &simpletable.Header{
			Cells: []*simpletable.Cell{
				{Align: simpletable.AlignCenter, Text: "#"},
				{Align: simpletable.AlignCenter, Text: "ID"},
				{Align: simpletable.AlignCenter, Text: "USERNAME"},
				{Align: simpletable.AlignCenter, Text: "PASSWORD"},
				{Align: simpletable.AlignCenter, Text: "URI"},
				{Align: simpletable.AlignCenter, Text: "DESCRIPTION"},
			},
		}
		for i, e := range creds {
			row := []*simpletable.Cell{
				{Align: simpletable.AlignLeft, Text: fmt.Sprintf("%d", i+1)},
				{Align: simpletable.AlignLeft, Text: e.ID},
				{Align: simpletable.AlignLeft, Text: e.Username},
				{Align: simpletable.AlignLeft, Text: e.Password},
				{Align: simpletable.AlignLeft, Text: e.Uri},
				{Align: simpletable.AlignLeft, Text: e.Description},
			}
			table.Body.Cells = append(table.Body.Cells, row)
		}
		localResources <- table
	}()

	go func() {
		wg.Wait()
		close(exit)
	}()

outerloop:
	for {
		select {
		case table := <-localResources:
			table.Println()
			fmt.Println()
		case e := <-errCh:
			err = multierror.Append(err, e)
		case <-exit:
			break outerloop
		}
	}

	return nil
}
