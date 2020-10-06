package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pavelmemory/faceit-users/internal"
	"github.com/pavelmemory/faceit-users/internal/storage"
)

func TestService_Create(t *testing.T) {
	userEntity := Entity{
		FirstName: "John",
		LastName:  "Doe",
		Nickname:  "johndoe",
		Email:     "johndoe@mail.com",
		Password:  "secret",
		Country:   "XX",
	}

	changeUserEntity := func(change func(entity *Entity)) Entity {
		user := userEntity
		change(&user)
		return user
	}

	t.Run("ok", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStorage := NewMockStorage(ctrl)
		mockStorage.EXPECT().
			Persist(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, run storage.Runner, user storage.User) (string, error) {
				require.Empty(t, user.ID)
				require.Equal(t, "John", user.FirstName)
				require.Equal(t, "Doe", user.LastName)
				require.Equal(t, "johndoe", user.Nickname)
				require.Equal(t, "johndoe@mail.com", user.Email)
				require.Equal(t, "secret", user.Password)
				require.Equal(t, "XX", user.Country)
				require.LessOrEqual(t, time.Now().Unix(), user.CreatedAt.Unix())
				require.LessOrEqual(t, time.Now().Unix(), user.UpdatedAt.Unix())
				return "1-2-3-4", nil
			})

		srv := NewService(testStorage{Transactioner: testTransactioner{}, Storage: mockStorage})
		id, err := srv.Create(Context(), userEntity)

		require.NoError(t, err)
		require.Equal(t, "1-2-3-4", id)
	})

	t.Run("can't save to storage", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStorage := NewMockStorage(ctrl)
		mockStorage.EXPECT().
			Persist(gomock.Any(), gomock.Any(), gomock.Any()).
			Return("", assert.AnError)

		srv := NewService(testStorage{Transactioner: testTransactioner{}, Storage: mockStorage})
		_, err := srv.Create(Context(), userEntity)

		require.Error(t, err)
		assert.True(t, errors.Is(err, assert.AnError))
		require.Contains(t, err.Error(), "persist user")
	})

	t.Run("validate", func(t *testing.T) {
		for title, tc := range map[string]struct {
			user    Entity
			details map[string]interface{}
		}{
			// TODO: only each type of check is covered, each field should have all verification of all checks
			"first name is empty": {
				user: changeUserEntity(func(entity *Entity) {
					entity.FirstName = ""
				}),
				details: map[string]interface{}{"FirstName": "blank or empty"},
			},
			"last name is blank": {
				user: changeUserEntity(func(entity *Entity) {
					entity.LastName = "  \t "
				}),
				details: map[string]interface{}{"LastName": "blank or empty"},
			},
			"nickname is too long": {
				user: changeUserEntity(func(entity *Entity) {
					entity.Nickname = "012345678901234567890123456789X"
				}),
				details: map[string]interface{}{"Nickname": "exceeds max length: 30"},
			},
			"invalid email": {
				user: changeUserEntity(func(entity *Entity) {
					entity.Email = "bad@mail"
				}),
				details: map[string]interface{}{"Email": "invalid format"},
			},
		} {
			t.Run(title, func(t *testing.T) {
				srv := NewService(nil)
				_, err := srv.Create(Context(), tc.user)
				require.Error(t, err)
				var verr ValidationError
				require.True(t, errors.As(err, &verr))
				require.Equal(t, verr.Cause, internal.ErrBadInput)
				require.Equal(t, tc.details, verr.Details)
			})
		}
	})
}

func Context() context.Context {
	return context.Background()
}

type testTransactioner struct{}

func (testTransactioner) WithTx(_ context.Context, call func(runner storage.Runner) error) error {
	return call(nil)
}

func (testTransactioner) WithoutTx(_ context.Context, call func(runner storage.Runner) error) error {
	return call(nil)
}

type testStorage struct {
	Transactioner
	Storage
}

func (ts testStorage) WithTx(ctx context.Context, call func(runner storage.Runner) error) error {
	return ts.Transactioner.WithTx(ctx, call)
}

func (ts testStorage) WithoutTx(ctx context.Context, call func(runner storage.Runner) error) error {
	return ts.Transactioner.WithoutTx(ctx, call)
}
