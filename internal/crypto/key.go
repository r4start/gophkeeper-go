package crypto

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/sha3"
)

const (
	SaltSize  = 64
	KeySize   = 64
	KeyRounds = 1000000
)

type Secret struct {
	Key  []byte
	Salt []byte
}

func GenerateSecretKey(baseSecret []byte, keySize, saltSize, rounds int) (*Secret, error) {
	salt := make([]byte, saltSize)
	read, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}

	if read < 64 {
		return nil, fmt.Errorf("not enough entropy")
	}

	key := pbkdf2.Key(baseSecret, salt, rounds, keySize, sha3.New512)

	return &Secret{
		Key:  key,
		Salt: salt,
	}, nil
}

func GenerateMasterKey(baseSecret []byte) (*Secret, error) {
	return GenerateSecretKey(baseSecret, KeySize, SaltSize, KeyRounds)
}

func RecoverMasterKey(baseSecret, salt []byte) (*Secret, error) {
	key := pbkdf2.Key(baseSecret, salt, KeyRounds, KeySize, sha3.New512)

	return &Secret{
		Key:  key,
		Salt: salt,
	}, nil
}
