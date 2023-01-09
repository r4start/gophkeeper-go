package app

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"time"

	"golang.org/x/crypto/sha3"

	"github.com/golang-jwt/jwt/v4"

	"github.com/r4start/goph-keeper/internal/server/storage"
)

const (
	tokenIssuer          = "gophkeeper"
	tokenAudience        = "token"
	refreshTokenAudience = "refresh"
	tokenLivenessPeriod  = time.Hour
	tokenSize            = 64
	keySize              = 64
)

var (
	_ Authorizer = (*authorizerImpl)(nil)

	signingMethod = jwt.SigningMethodHS512

	ErrBadCredentials     = errors.New("bad credentials")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrBadSignMethod      = errors.New("bad sign method")
	ErrExpiredToken       = errors.New("expired token")
	ErrInvalidToken       = errors.New("invalid token")
)

type AuthData struct {
	ID           string
	Token        string
	RefreshToken string
	ExpiresAt    int64
	KeySalt      []byte
}

type Authorizer interface {
	Register(ctx context.Context, login, password string, salt []byte) (*AuthData, error)
	Authorize(ctx context.Context, login, password string) (*AuthData, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthData, error)
	IsValidToken(ctx context.Context, token string) error
	AuthorizeWithToken(ctx context.Context, token string) (*AuthData, error)
}

type jwtClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type authorizerImpl struct {
	userService   storage.UserService
	signKey       []byte
	signingMethod jwt.SigningMethod
}

func NewAuthorizer(us storage.UserService, key []byte) (*authorizerImpl, error) {
	if len(key) != keySize {
	}

	return &authorizerImpl{
		userService:   us,
		signKey:       key,
		signingMethod: signingMethod,
	}, nil
}

func (a *authorizerImpl) Register(ctx context.Context, login, password string, keySalt []byte) (*AuthData, error) {
	if len(login) == 0 || len(password) == 0 || len(keySalt) == 0 {
		return nil, ErrBadCredentials
	}

	secret, salt, err := generateSecret(password, nil)
	if err != nil {
		return nil, err
	}

	id, err := a.userService.Add(ctx, login, keySalt, salt, secret)
	if err != nil {
		return nil, err
	}

	userID := id.String()
	auth, err := a.generateAuthData(userID)
	if err != nil {
		return nil, err
	}

	auth.ID = userID
	return auth, nil
}

func (a *authorizerImpl) Authorize(ctx context.Context, login, password string) (*AuthData, error) {
	if len(login) == 0 || len(password) == 0 {
		return nil, ErrBadCredentials
	}

	u, err := a.userService.GetByLogin(ctx, login)
	if err != nil {
		return nil, err
	}

	generatedSecret, _, err := generateSecret(password, u.Salt)
	if err != nil {
		return nil, err
	}

	if subtle.ConstantTimeCompare(generatedSecret, u.Secret) == 0 {
		return nil, ErrInvalidCredentials
	}

	auth, err := a.generateAuthData(u.ID.String())
	if err != nil {
		return nil, err
	}
	auth.ID = u.ID.String()
	auth.KeySalt = u.KeySalt

	return auth, nil
}

func (a *authorizerImpl) RefreshToken(ctx context.Context, refreshToken string) (*AuthData, error) {
	return nil, errors.New("unimplemented")
}

func (a *authorizerImpl) IsValidToken(_ context.Context, token string) error {
	claims := &jwtClaims{}
	t, err := a.parseToken(token, claims)
	if err != nil {
		return err
	}
	return a.checkToken(t, claims)
}

func (a *authorizerImpl) AuthorizeWithToken(ctx context.Context, token string) (*AuthData, error) {
	claims := &jwtClaims{}
	t, err := a.parseToken(token, claims)
	if err != nil {
		return nil, err
	}
	if err := a.checkToken(t, claims); err != nil {
		return nil, err
	}

	user, err := a.userService.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if user.IsDeleted {
		return nil, errors.New("unauthenticated")
	}

	return &AuthData{
		ID:           claims.UserID,
		Token:        token,
		RefreshToken: "",
		ExpiresAt:    claims.ExpiresAt.Unix(),
		KeySalt:      user.KeySalt,
	}, nil
}

func (a *authorizerImpl) generateAuthData(uid string) (*AuthData, error) {
	ts := time.Now().UTC()

	outputToken, err := createSignedToken(a.signingMethod, ts, tokenSize, tokenAudience, uid, a.signKey)
	if err != nil {
		return nil, err
	}

	outRefreshToken, err := createSignedToken(a.signingMethod, ts, tokenSize, refreshTokenAudience, uid, a.signKey)
	if err != nil {
		return nil, err
	}

	return &AuthData{
		Token:        *outputToken,
		RefreshToken: *outRefreshToken,
		ExpiresAt:    ts.Add(tokenLivenessPeriod).Unix(),
	}, nil
}

func (a *authorizerImpl) parseToken(token string, claims *jwtClaims) (*jwt.Token, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrBadSignMethod
		}

		return a.signKey, nil
	}

	return jwt.ParseWithClaims(token, claims, keyFunc)
}

func (a *authorizerImpl) checkToken(token *jwt.Token, claims *jwtClaims) error {
	if !token.Valid {
		return ErrInvalidToken
	}

	now := time.Now().UTC()

	if claims.ExpiresAt.Before(now) {
		return ErrExpiredToken
	}

	if claims.NotBefore.After(now) {
		return ErrInvalidToken
	}

	return nil
}

func createToken(method jwt.SigningMethod, ts time.Time, tokenSize int, aud, uid string) (*jwt.Token, error) {
	tokenID := make([]byte, tokenSize)
	if _, err := rand.Read(tokenID); err != nil {
		return nil, err
	}

	claims := &jwtClaims{
		uid,
		jwt.RegisteredClaims{
			Issuer:    tokenIssuer,
			Audience:  []string{aud},
			ExpiresAt: jwt.NewNumericDate(ts.Add(tokenLivenessPeriod)),
			NotBefore: jwt.NewNumericDate(ts),
			ID:        base64.URLEncoding.EncodeToString(tokenID),
		}}

	return jwt.NewWithClaims(method, claims), nil
}

func createSignedToken(method jwt.SigningMethod, ts time.Time, tokenSize int, aud, uid string, signKey []byte) (*string, error) {
	token, err := createToken(method, ts, tokenSize, aud, uid)
	if err != nil {
		return nil, err
	}

	signedToken, err := token.SignedString(signKey)
	if err != nil {
		return nil, err
	}

	return &signedToken, nil
}

func generateSecret(s string, salt []byte) ([]byte, []byte, error) {
	hasher := sha3.New512()

	if salt == nil {
		salt = make([]byte, hasher.BlockSize())
		if _, err := rand.Read(salt); err != nil {
			return nil, nil, err
		}
	}

	hasher.Write(salt)
	hasher.Write([]byte(s))
	secret := hasher.Sum(nil)
	return secret, salt, nil
}
