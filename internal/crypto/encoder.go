package crypto

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/sha3"

	"github.com/r4start/goph-keeper/internal/client/resource"
)

var (
	_ resource.Encoder = (*aesGCMEncoder)(nil)
	_ resource.Decoder = (*aesGCMEncoder)(nil)
)

const (
	nonceSize = 12
	keySize   = 32
)

type aesGCMEncoder struct {
	bc      cipher.Block
	aead    cipher.AEAD
	key     []byte
	salt    []byte
	nonce   []byte
	wr      io.Writer
	written int
}

func NewAesGcmEncoder(masterKey []byte) (*aesGCMEncoder, error) {
	hash := sha3.New512

	salt := make([]byte, hash().Size())
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	h := hkdf.New(hash, masterKey, salt, nil)
	key := make([]byte, keySize)
	if _, err := io.ReadFull(h, key); err != nil {
		return nil, err
	}

	nonce := make([]byte, nonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	bc, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(bc)
	if err != nil {
		return nil, err
	}

	return &aesGCMEncoder{
		bc:    bc,
		aead:  aesgcm,
		key:   key,
		salt:  salt,
		nonce: nonce,
	}, nil
}

func RestoreAesGcmEncoder(masterKey, salt []byte) (*aesGCMEncoder, error) {
	hash := sha3.New512

	if len(salt) != hash().Size() {
		return nil, fmt.Errorf("bad salt size: got %d; need %d", len(salt), hash().Size())
	}

	h := hkdf.New(hash, masterKey, salt, nil)
	key := make([]byte, keySize)
	if _, err := io.ReadFull(h, key); err != nil {
		return nil, err
	}

	nonce := make([]byte, nonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	bc, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(bc)
	if err != nil {
		return nil, err
	}

	return &aesGCMEncoder{
		bc:    bc,
		aead:  aesgcm,
		key:   key,
		salt:  salt,
		nonce: nonce,
	}, nil
}

func (a *aesGCMEncoder) Encode(ctx context.Context, data []byte) ([]byte, error) {
	nonceLen := uint64(len(a.nonce))
	if nonceLen == 0 || len(a.key) == 0 {
		return nil, fmt.Errorf("bad encryption parameters: %d key len; %d nonce len", len(a.key), nonceLen)
	}

	res := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(res, nonceLen)
	res = res[:n]
	res = append(res, a.nonce...)
	res = append(res, a.aead.Seal(nil, a.nonce, data, nil)...)
	return res, nil
}

func (a *aesGCMEncoder) Decode(ctx context.Context, data []byte) ([]byte, error) {
	nonceLen, n := binary.Uvarint(data)
	if n < 1 {
		return nil, fmt.Errorf("data format is invalid")
	}
	buf := data[n:]
	nonce := buf[:nonceLen]
	buf = buf[nonceLen:]
	return a.aead.Open(nil, nonce, buf, nil)
}

func (a *aesGCMEncoder) Salt() []byte {
	return a.salt
}

func (a *aesGCMEncoder) Key() []byte {
	return a.key
}
