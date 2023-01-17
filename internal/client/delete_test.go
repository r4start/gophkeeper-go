package client

import (
	"context"
	"testing"

	"github.com/r4start/goph-keeper/internal/client/storage"
	"github.com/stretchr/testify/assert"
)

func TestDeleter_DeleteTestDeleter_Delete(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()
	st := storage.NewMockStorage()
	client := newMockClient()
	up := NewUploader(client, st, tempDir)

	card := storage.CardData{
		Name:         "Test card",
		Number:       "5555 5555 5555 5555",
		Holder:       "Tririr Eritndcxh",
		ExpiryDate:   "11/22",
		SecurityCode: "111",
	}

	assert.NoError(t, up.UploadCard(ctx, card))
	assert.Equal(t, 1, len(st.Cards))

	cred := storage.CredentialData{
		Username:    "uu1",
		Password:    "sjksjs",
		Uri:         "snshjs",
		Description: "dsjdsjd",
	}

	assert.NoError(t, up.UploadCredentials(ctx, cred))
	assert.Equal(t, 1, len(st.Creds))

	tmpDir := t.TempDir()
	testFile, err := createTestFile(tmpDir)
	assert.NoError(t, err)

	assert.NoError(t, up.UploadFiles(ctx, []string{testFile}))
	assert.Equal(t, 1, len(st.Files))

	cards := make([]string, 0, len(st.Cards))
	for k := range st.Cards {
		cards = append(cards, k)
	}

	deleter := NewDeleter(client, st)
	assert.NoError(t, deleter.Delete(ctx, cards))
	assert.Empty(t, st.Cards)

	creds := make([]string, 0, len(st.Creds))
	for k := range st.Creds {
		creds = append(creds, k)
	}

	assert.NoError(t, deleter.Delete(ctx, creds))
	assert.Empty(t, st.Creds)

	files := make([]string, 0, len(st.Files))
	for k := range st.Files {
		files = append(files, k)
	}

	assert.NoError(t, deleter.Delete(ctx, files))
	assert.Empty(t, st.Files)
}
