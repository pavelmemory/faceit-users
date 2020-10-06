package webhttp

import (
	"context"
	"net/http"
	"net/url"

	"github.com/go-chi/chi"

	"github.com/pavelmemory/faceit-users/internal/logging"
	"github.com/pavelmemory/faceit-users/internal/user"
)

//go:generate mockgen -source=user.go -destination mock.go -package webhttp UserService

// UserService provides set of operations available to operate on the user entity.
type UserService interface {
	// Create creates a new user entity and returns its unique identifier.
	Create(ctx context.Context, user user.Entity) (string, error)
	// Get returns user entity by its unique identifier.
	Get(ctx context.Context, id string) (user.Entity, error)
	// Update updates user entity found by unique identifier.
	Update(ctx context.Context, id string, user user.Entity) error
	// Delete removes user entity by its unique identifier.
	Delete(ctx context.Context, id string) error
}

// NewUsersHandler returns HTTP handler initialized with provided service abstraction.
func NewUsersHandler(userService UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// UserHandler handles request for the user entity(-ies).
type UserHandler struct {
	userService UserService
	mapper      Mapper
}

// Register creates a binding between method handlers and endpoints.
func (uh *UserHandler) Register(router chi.Router) {
	router = router.With(LogRequest())
	router.With(ProducesJSON, AcceptsJSON).Method(http.MethodPost, uh.urlPrefix(), http.HandlerFunc(uh.Create))
	router.With(ProducesJSON).Method(http.MethodGet, uh.urlPrefix()+"/{id}", http.HandlerFunc(uh.Get))
	router.With(AcceptsJSON).Method(http.MethodPut, uh.urlPrefix()+"/{id}", http.HandlerFunc(uh.Update))
	router.Method(http.MethodDelete, uh.urlPrefix()+"/{id}", http.HandlerFunc(uh.Delete))
}

func (uh *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := uh.logger(ctx, "Create")

	logger.Debug("start")
	defer logger.Debug("end")

	var req CreateUserReq
	if err := Decode(r.Body, &req); err != nil {
		logger.WithError(err).Error("decode payload")
		ErrorResponse{Cause: err, StatusCode: http.StatusBadRequest}.Write(logger, w)
		return
	}

	id, err := uh.userService.Create(ctx, uh.mapper.createUserReq2Entity(req))
	if err != nil {
		logger.WithError(err).Error("creation of the Booking")
		WriteError(w, logger, err)
		return
	}

	w.Header().Set("location", uh.urlPrefix()+"/"+url.PathEscape(id))
	w.WriteHeader(http.StatusCreated)
}

func (uh *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := uh.logger(ctx, "Get")

	logger.Debug("start")
	defer logger.Debug("end")

	id := uh.pathParam(r, "id")
	u, err := uh.userService.Get(ctx, id)
	if err != nil {
		logger.WithError(err).WithString("id", id).Error("get user by id")
		WriteError(w, logger, err)
		return
	}

	if err := Encode(w, uh.mapper.entity2GetUserResp(u)); err != nil {
		logger.WithError(err).Error("encode entity")
		ErrorResponse{Cause: err, StatusCode: http.StatusInternalServerError}.Write(logger, w)
		return
	}
}

func (uh *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := uh.logger(ctx, "Update")

	logger.Debug("start")
	defer logger.Debug("end")

	id := uh.pathParam(r, "id")

	var req UpdateUserReq
	if err := Decode(r.Body, &req); err != nil {
		logger.WithError(err).Error("decode payload")
		ErrorResponse{Cause: err, StatusCode: http.StatusBadRequest}.Write(logger, w)
		return
	}

	if err := uh.userService.Update(ctx, id, uh.mapper.updateUserReq2Entity(req)); err != nil {
		logger.WithError(err).WithString("id", id).Error("update user")
		WriteError(w, logger, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (uh *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := uh.logger(ctx, "Delete")

	logger.Debug("start")
	defer logger.Debug("end")

	id := uh.pathParam(r, "id")

	if err := uh.userService.Delete(ctx, id); err != nil {
		logger.WithError(err).WithString("id", id).Error("delete user")
		WriteError(w, logger, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (uh *UserHandler) urlPrefix() string {
	return "/users"
}

func (uh *UserHandler) pathParam(r *http.Request, name string) string {
	return chi.URLParam(r, name)
}

func (uh *UserHandler) logger(ctx context.Context, method string) logging.Logger {
	return logging.FromContext(ctx).WithString("component", "UserHandler").WithString("method", method)
}
