package client

import (
	"context"
	"crypto/rand"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/r4start/goph-keeper/internal/client/storage"
)

func TestUploader_UploadCard(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()
	st := storage.NewMockStorage()
	up := NewUploader(newMockClient(), st, tempDir)

	card := storage.CardData{
		Name:         "Test card",
		Number:       "5555 5555 5555 5555",
		Holder:       "Tririr Eritndcxh",
		ExpiryDate:   "11/22",
		SecurityCode: "111",
	}

	assert.NoError(t, up.UploadCard(ctx, card))
	assert.Equal(t, 1, len(st.Cards))

	for _, v := range st.Cards {
		assert.Equal(t, card.Name, v.Name)
		assert.Equal(t, card.Number, v.Number)
		assert.Equal(t, card.Holder, v.Holder)
		assert.Equal(t, card.ExpiryDate, v.ExpiryDate)
		assert.Equal(t, card.SecurityCode, v.SecurityCode)
	}
}

func TestUploader_UploadCredentials(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()
	st := storage.NewMockStorage()
	up := NewUploader(newMockClient(), st, tempDir)

	cred := storage.CredentialData{
		Username:    "uu1",
		Password:    "sjksjs",
		Uri:         "snshjs",
		Description: "dsjdsjd",
	}

	assert.NoError(t, up.UploadCredentials(ctx, cred))
	assert.Equal(t, 1, len(st.Creds))
	for _, v := range st.Creds {
		assert.Equal(t, cred.Username, v.Username)
		assert.Equal(t, cred.Password, v.Password)
		assert.Equal(t, cred.Uri, v.Uri)
		assert.Equal(t, cred.Description, v.Description)
	}
}

func TestUploader_UploadFiles(t *testing.T) {
	ctx := context.Background()
	syncDir := t.TempDir()
	st := storage.NewMockStorage()
	up := NewUploader(newMockClient(), st, syncDir)

	tmpDir := t.TempDir()
	testFile, err := createTestFile(tmpDir)
	assert.NoError(t, err)

	assert.NoError(t, up.UploadFiles(ctx, []string{testFile}))
	assert.Equal(t, 1, len(st.Files))
}

func createTestFile(tmpDir string) (string, error) {
	buf := make([]byte, 111*1024*1024)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	fileName := tmpDir + string(os.PathSeparator) + "test_files"
	f, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = f.Close()
	}()

	_, err = f.Write(buf)
	if err != nil {
		return "", err
	}

	return fileName, nil
}
