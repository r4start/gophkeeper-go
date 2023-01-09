package storage

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/google/uuid"
)

type MockStorage struct {
	User  UserData
	Cards map[string]CardData
	Creds map[string]CredentialData
	Files map[string]FileData
}

func NewMockStorage() *MockStorage {
	mk, _ := generateRandom(64)
	salt, _ := generateRandom(64)
	user, _ := uuid.NewRandom()

	ud := UserData{
		UserID:    user.String(),
		MasterKey: mk,
		Salt:      salt,
	}
	return &MockStorage{
		User:  ud,
		Cards: make(map[string]CardData),
		Creds: make(map[string]CredentialData),
		Files: make(map[string]FileData),
	}
}

func (ms *MockStorage) AddCard(ctx context.Context, data *CardData) error {
	if _, ok := ms.Cards[data.ID]; !ok {
		ms.Cards[data.ID] = *data
		return nil
	}
	return fmt.Errorf("duplicate entry")
}

func (ms *MockStorage) AddCredentials(ctx context.Context, data *CredentialData) error {
	if _, ok := ms.Creds[data.ID]; !ok {
		ms.Creds[data.ID] = *data
		return nil
	}
	return fmt.Errorf("duplicate entry")
}

func (ms *MockStorage) AddFile(ctx context.Context, fd *FileData) error {
	if _, ok := ms.Files[fd.ID]; !ok {
		ms.Files[fd.ID] = *fd
		return nil
	}
	return fmt.Errorf("duplicate entry")
}

func (*MockStorage) CardData(ctx context.Context, id string) (*CardData, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (*MockStorage) CredentialData(ctx context.Context, id string) (*CredentialData, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (ms *MockStorage) DeleteCard(ctx context.Context, id string) error {
	delete(ms.Cards, id)
	return nil
}

func (ms *MockStorage) DeleteCredential(ctx context.Context, id string) error {
	delete(ms.Creds, id)
	return nil
}

func (ms *MockStorage) DeleteFile(ctx context.Context, fd *FileData) error {
	delete(ms.Files, fd.ID)
	return nil
}

func (ms *MockStorage) FileData(ctx context.Context, id string) (*FileData, error) {
	data, ok := ms.Files[id]
	if !ok {
		return nil, fmt.Errorf("no file with id:%s", id)
	}
	return &data, nil
}

func (ms *MockStorage) ListCards(ctx context.Context) ([]CardData, error) {
	result := make([]CardData, 0, len(ms.Cards))
	for _, v := range ms.Cards {
		result = append(result, v)
	}
	return result, nil
}

func (ms *MockStorage) ListCredentials(ctx context.Context) ([]CredentialData, error) {
	result := make([]CredentialData, 0, len(ms.Creds))
	for _, v := range ms.Creds {
		result = append(result, v)
	}
	return result, nil
}

func (ms *MockStorage) ListFiles(ctx context.Context) ([]FileData, error) {
	result := make([]FileData, 0, len(ms.Files))
	for _, v := range ms.Files {
		result = append(result, v)
	}
	return result, nil
}

func (*MockStorage) SetUserData(ctx context.Context, ud *UserData) error {
	return nil
}

func (ms *MockStorage) UserData(ctx context.Context) (*UserData, error) {
	return &ms.User, nil
}

func generateRandom(size int) ([]byte, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	return buf, nil
}
