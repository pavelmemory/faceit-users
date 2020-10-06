package webhttp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pavelmemory/faceit-users/internal/user"
	"github.com/stretchr/testify/require"

	"github.com/pavelmemory/faceit-users/internal/logging"
)

func TestUserHandler_Create(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		logger := logging.NewTestLogger()
		r := NewRouter(logger)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUserService := NewMockUserService(ctrl)
		mockUserService.EXPECT().Create(gomock.Any(), gomock.Any()).Return("1-2-3-4", nil)

		userHandler := NewUsersHandler(mockUserService)
		userHandler.Register(r)

		req := httptest.NewRequest(http.MethodPost, "http://localhost/users", strings.NewReader(`{}`))
		req.Header.Set("content-type", "application/json")
		resp := httptest.NewRecorder()

		r.ServeHTTP(resp, req)

		require.Equal(t, http.StatusCreated, resp.Code)
		require.Equal(t, "/users/1-2-3-4", resp.Header().Get("location"))
	})

	// TODO: other scenarios of input as well as response from the 'mockUserService'
}

func TestUserHandler_Get(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		logger := logging.NewTestLogger()
		r := NewRouter(logger)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUserService := NewMockUserService(ctrl)
		mockUserService.EXPECT().Get(gomock.Any(), "1-2-3-4").Return(user.Entity{FirstName: "fn"}, nil)

		userHandler := NewUsersHandler(mockUserService)
		userHandler.Register(r)

		req := httptest.NewRequest(http.MethodGet, "http://localhost/users/1-2-3-4", nil)
		resp := httptest.NewRecorder()

		r.ServeHTTP(resp, req)

		require.Equal(t, http.StatusOK, resp.Code)
		require.JSONEq(t, `{"first_name":"fn"}`, resp.Body.String())
	})

	// TODO: other scenarios of input as well as response from the 'mockUserService'
}

// TODO: other endpoints should be covered as well
