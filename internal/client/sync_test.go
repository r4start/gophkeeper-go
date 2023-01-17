package client

import (
	"context"
	"testing"

	"github.com/r4start/goph-keeper/internal/client/storage"
	"github.com/stretchr/testify/assert"
)

func TestSynchronizer_SyncFile(t *testing.T) {
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

	st.Files = make(map[string]storage.FileData)

	sync := NewSynchronizer(client, st, tempDir)
	assert.NoError(t, sync.Sync(ctx))
	assert.Equal(t, 1, len(st.Files))
}

func TestSynchronizer_SyncCard(t *testing.T) {
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

	st.Cards = make(map[string]storage.CardData)

	sync := NewSynchronizer(client, st, tempDir)
	assert.NoError(t, sync.Sync(ctx))
	assert.Equal(t, 1, len(st.Cards))
}

func TestSynchronizer_SyncCred(t *testing.T) {
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

	st.Creds = make(map[string]storage.CredentialData)

	sync := NewSynchronizer(client, st, tempDir)
	assert.NoError(t, sync.Sync(ctx))
	assert.Equal(t, 1, len(st.Creds))
}
