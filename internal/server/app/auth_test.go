package app

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/r4start/goph-keeper/internal/server/storage"
)

var (
	errUserAlreadyExist = errors.New("user already exist")
	errUserNotFound     = errors.New("user not found")
)

func generateToken(t *testing.T, uid string, key []byte, ts time.Time) string {
	outputToken, err := createSignedToken(signingMethod, ts, tokenSize, tokenAudience, uid, key)
	assert.NoError(t, err)
	return *outputToken
}

func generateKey(size int) ([]byte, error) {
	tokenID := make([]byte, size)
	if _, err := rand.Read(tokenID); err != nil {
		return nil, err
	}
	return tokenID, nil
}

type authData struct {
	User    *storage.UserID
	Secret  []byte
	Salt    []byte
	KeySalt []byte
}

type mockUserService struct {
	Users *sync.Map
}

func (m *mockUserService) Add(_ context.Context, login string, keySalt, salt, secret []byte) (*storage.UserID, error) {
	userID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	u := storage.UserID(userID)
	value, loaded := m.Users.LoadOrStore(login, authData{
		User:    &u,
		Secret:  secret,
		Salt:    salt,
		KeySalt: keySalt,
	})

	if loaded {
		return nil, errUserAlreadyExist
	}

	return value.(authData).User, nil
}

func (m *mockUserService) GetByLogin(_ context.Context, login string) (*storage.User, error) {
	u, loaded := m.Users.Load(login)
	if !loaded {
		return nil, errUserNotFound
	}

	auth := u.(authData)
	return &storage.User{
		ID:        *auth.User,
		Login:     login,
		Salt:      auth.Salt,
		Secret:    auth.Secret,
		IsDeleted: false,
	}, nil
}

func (m *mockUserService) GetByID(ctx context.Context, id string) (*storage.User, error) {
	//TODO implement me
	panic("implement me")
}

func (m *mockUserService) Close() error {
	return nil
}

func NewMockUserService() *mockUserService {
	return &mockUserService{Users: new(sync.Map)}
}

func Test_authorizerImpl_Register(t *testing.T) {
	type args struct {
		login    string
		password string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Register #1",
			args: args{
				login:    "t1",
				password: "t1",
			},
			wantErr: false,
		},
		{
			name: "Register #2",
			args: args{
				login:    "t2",
				password: "",
			},
			wantErr: true,
		},
		{
			name: "Register #3",
			args: args{
				login:    "",
				password: "t3",
			},
			wantErr: true,
		},
		{
			name: "Register #4",
			args: args{
				login:    "",
				password: "",
			},
			wantErr: true,
		},
	}

	signKey, err := generateKey(keySize)
	assert.NoError(t, err)

	usersStorage := NewMockUserService()
	ctx := context.Background()

	keySalt := make([]byte, 64)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := NewAuthorizer(usersStorage, signKey)
			assert.NoError(t, err)
			got, err := a.Register(ctx, tt.args.login, tt.args.password, keySalt)
			if !tt.wantErr {
				assert.NoError(t, err)

				keyFunc := func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
					}

					return signKey, nil
				}

				claims := &jwt.RegisteredClaims{}
				token, err := jwt.ParseWithClaims(got.Token, claims, keyFunc)
				assert.NoError(t, err)
				assert.True(t, token.Valid)

				hasCorrectAudience := false
				for _, aud := range claims.Audience {
					if aud == tokenAudience {
						hasCorrectAudience = true
						break
					}
				}
				assert.True(t, hasCorrectAudience)

				token, err = jwt.ParseWithClaims(got.RefreshToken, claims, keyFunc)
				assert.NoError(t, err)
				assert.True(t, token.Valid)

				hasCorrectAudience = false
				for _, aud := range claims.Audience {
					if aud == refreshTokenAudience {
						hasCorrectAudience = true
						break
					}
				}
				assert.True(t, hasCorrectAudience)

				auth, ok := usersStorage.Users.Load(tt.args.login)
				assert.True(t, ok)
				assert.NotZero(t, len(auth.(authData).Salt))
				assert.NotZero(t, len(auth.(authData).Secret))
			} else {
				assert.Error(t, err)
			}
		})
	}

	a, err := NewAuthorizer(usersStorage, signKey)
	assert.NoError(t, err)
	_, err = a.Register(ctx, tests[0].args.login, tests[0].args.password, keySalt)
	assert.Error(t, err)

	samePassLogin := "samePassTest"
	_, err = a.Register(ctx, samePassLogin, tests[0].args.password, keySalt)
	assert.NoError(t, err)

	authOrigin, ok := usersStorage.Users.Load(tests[0].args.login)
	assert.True(t, ok)

	authOther, ok := usersStorage.Users.Load(samePassLogin)
	assert.True(t, ok)

	assert.NotEqual(t, authOrigin.(authData).Salt, authOther.(authData).Salt)
	assert.NotEqual(t, authOrigin.(authData).Secret, authOther.(authData).Secret)
}

func Test_authorizerImpl_Authorize(t *testing.T) {
	type args struct {
		login    string
		password string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Authorize #1",
			args: args{
				login:    "t1",
				password: "t1",
			},
			wantErr: false,
		},
		{
			name: "Authorize #2",
			args: args{
				login:    "t2",
				password: "t2",
			},
			wantErr: false,
		},
		{
			name: "Authorize #3",
			args: args{
				login:    "",
				password: "",
			},
			wantErr: true,
		},
		{
			name: "Authorize #4",
			args: args{
				login:    "t4",
				password: "",
			},
			wantErr: true,
		},
		{
			name: "Authorize #5",
			args: args{
				login:    "",
				password: "t5",
			},
			wantErr: true,
		},
		{
			name: "Authorize #6",
			args: args{
				login:    "t6",
				password: "t6",
			},
			wantErr: true,
		},
	}

	signKey, err := generateKey(keySize)
	assert.NoError(t, err)

	keySalt := make([]byte, 64)

	usersStorage := NewMockUserService()
	ctx := context.Background()

	a, err := NewAuthorizer(usersStorage, signKey)
	assert.NoError(t, err)

	_, err = a.Register(ctx, tests[0].args.login, tests[0].args.password, keySalt)
	assert.NoError(t, err)

	_, err = a.Register(ctx, tests[1].args.login, tests[1].args.password, keySalt)
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := a.Authorize(ctx, tt.args.login, tt.args.password)
			if !tt.wantErr {
				assert.NoError(t, err)

				keyFunc := func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
					}

					return signKey, nil
				}

				claims := &jwt.RegisteredClaims{}
				token, err := jwt.ParseWithClaims(got.Token, claims, keyFunc)
				assert.NoError(t, err)
				assert.True(t, token.Valid)

				hasCorrectAudience := false
				for _, aud := range claims.Audience {
					if aud == tokenAudience {
						hasCorrectAudience = true
						break
					}
				}
				assert.True(t, hasCorrectAudience)

				token, err = jwt.ParseWithClaims(got.RefreshToken, claims, keyFunc)
				assert.NoError(t, err)
				assert.True(t, token.Valid)

				hasCorrectAudience = false
				for _, aud := range claims.Audience {
					if aud == refreshTokenAudience {
						hasCorrectAudience = true
						break
					}
				}
				assert.True(t, hasCorrectAudience)
			} else {
				assert.Error(t, err)
			}
		})
	}

	_, err = a.Authorize(ctx, tests[0].args.login, tests[1].args.login)
	assert.Error(t, err)
}

func Test_authorizerImpl_IsValidToken(t *testing.T) {
	userID, err := uuid.NewRandom()
	assert.NoError(t, err)
	uid := userID.String()

	signKey, err := generateKey(keySize)
	assert.NoError(t, err)

	tests := []struct {
		name    string
		token   string
		isValid bool
	}{
		{
			name:    "Empty Token",
			token:   "",
			isValid: false,
		},
		{
			name:    "Valid Token",
			token:   generateToken(t, uid, signKey, time.Now().UTC()),
			isValid: true,
		},
		{
			name:    "Expired",
			token:   generateToken(t, uid, signKey, time.Now().UTC().Add(-2*tokenLivenessPeriod)),
			isValid: false,
		},
		{
			name:    "Before valid",
			token:   generateToken(t, uid, signKey, time.Now().UTC().Add(2*tokenLivenessPeriod)),
			isValid: false,
		},
	}

	usersStorage := NewMockUserService()
	ctx := context.Background()
	a, err := NewAuthorizer(usersStorage, signKey)
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := a.IsValidToken(ctx, tt.token)
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func Test_authorizerImpl_RefreshToken(t *testing.T) {
	type fields struct {
		userService   storage.UserService
		signKey       []byte
		signingMethod jwt.SigningMethod
	}
	type args struct {
		ctx          context.Context
		refreshToken string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *AuthData
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &authorizerImpl{
				userService:   tt.fields.userService,
				signKey:       tt.fields.signKey,
				signingMethod: tt.fields.signingMethod,
			}
			got, err := a.RefreshToken(tt.args.ctx, tt.args.refreshToken)
			if !tt.wantErr(t, err, fmt.Sprintf("RefreshToken(%v, %v)", tt.args.ctx, tt.args.refreshToken)) {
				return
			}
			assert.Equalf(t, tt.want, got, "RefreshToken(%v, %v)", tt.args.ctx, tt.args.refreshToken)
		})
	}
}
