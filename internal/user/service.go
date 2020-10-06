package user

import (
	"context"
	"fmt"
	"time"

	"github.com/pavelmemory/faceit-users/internal/storage"
)

type Entity struct {
	FirstName string
	LastName  string
	Nickname  string
	Email     string
	Password  string
	Country   string
}

//go:generate mockgen -source=service.go -destination mock.go -package user Storage

// Transactioner executes statements with/without explicitly open transaction.
type Transactioner interface {
	// WithTx executes provided callback inside of the transaction.
	// If callback returns an error the transaction will be rolled back, otherwise it will be committed.
	WithTx(context.Context, func(runner storage.Runner) error) error
	// WithoutTx executes provided callback without explicitly open transaction.
	WithoutTx(context.Context, func(runner storage.Runner) error) error
}

// Storage is a persistence storage for the user entity.
type Storage interface {
	Transactioner
	// Persist saves the user and returns it's unique generated ID.
	// It returns an error in case email or nickname is not unique.
	Persist(ctx context.Context, run storage.Runner, user storage.User) (string, error)
	// Retrieve returns user by supplied 'id'.
	// If user doesn't exist it returns an error.
	Retrieve(ctx context.Context, run storage.Runner, id string, forUpdate bool) (storage.User, error)
	// Update updates properties of the existing user and returns entity with old values.
	// If user doesn't exist it returns an error.
	Update(ctx context.Context, runner storage.Runner, id string, user storage.User) (storage.User, error)
	// Delete deletes user entity by its identifier.
	Delete(ctx context.Context, runner storage.Runner, id string) error
}

// NewService returns initialized user service.
func NewService(storage Storage) *Service {
	return &Service{storage: storage}
}

// Service allows to CRUD user entity.
// On each user modification it sends a notification about changes made to user entity.
type Service struct {
	storage Storage
}

// Create creates a new user entity and returns back its unique ID.
func (s *Service) Create(ctx context.Context, user Entity) (string, error) {
	if err := s.validate(user, propertyFirstName, propertyLastName, propertyNickname, propertyEmail, propertyCountry, propertyPassword); err != nil {
		return "", err
	}

	var id string
	if err := s.storage.WithTx(ctx, func(runner storage.Runner) (err error) {
		now := time.Now().UTC()
		newUser := entityUser(user)
		newUser.Password = user.Password
		newUser.CreatedAt = now.UTC()
		newUser.UpdatedAt = now.UTC()

		id, err = s.storage.Persist(ctx, runner, newUser)
		if err == nil {
			// TODO: send notification about creation of the user
		}
		return err
	}); err != nil {
		return "", fmt.Errorf("persist user: %w", err)
	}

	return id, nil
}

func (s *Service) validate(user Entity, validateProperties ...validationProperty) error {
	var validations []func() error
	for _, validateProperty := range validateProperties {
		var validation func() error

		switch validateProperty {
		case propertyFirstName:
			validation = validateBlankOrEmptyWithMaxLen(user.FirstName, validateProperty.String(), 50)
		case propertyLastName:
			validation = validateBlankOrEmptyWithMaxLen(user.LastName, validateProperty.String(), 50)
		case propertyNickname:
			validation = validateBlankOrEmptyWithMaxLen(user.Nickname, validateProperty.String(), 30)
		case propertyEmail:
			validation = validateEmailFormat(user.Email, validateProperty.String())
		case propertyPassword:
			validation = validateBlankOrEmptyWithMaxLen(user.Password, validateProperty.String(), 20)
		case propertyCountry:
			validation = validateBlankOrEmptyWithMaxLen(user.Country, validateProperty.String(), 2)
		default:
			continue
		}

		validations = append(validations, validation)
	}

	for _, validation := range validations {
		if err := validation(); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) Get(ctx context.Context, id string) (Entity, error) {
	var u storage.User
	if err := s.storage.WithoutTx(ctx, func(runner storage.Runner) (err error) {
		u, err = s.storage.Retrieve(ctx, runner, id, false)
		return
	}); err != nil {
		return Entity{}, fmt.Errorf("retrieve user %q: %w", id, err)
	}

	return userEntity(u), nil
}

func (s *Service) Update(ctx context.Context, id string, user Entity) error {
	if err := s.validate(user, propertyFirstName, propertyLastName, propertyNickname, propertyEmail, propertyCountry); err != nil {
		return err
	}

	changes := Changes{}
	err := s.storage.WithTx(ctx, func(runner storage.Runner) error {
		newUser := entityUser(user)
		newUser.UpdatedAt = time.Now().UTC()

		oldUser, err := s.storage.Update(ctx, runner, id, newUser)
		if err != nil {
			return fmt.Errorf("update user %q: %w", id, err)
		}

		// TODO: could be done via reflection magic, generated code, mapping lib, etc.
		if oldUser.FirstName != newUser.FirstName {
			changes.Add("FirstName", oldUser.FirstName, newUser.FirstName)
			oldUser.FirstName = newUser.FirstName
		}

		if oldUser.LastName != newUser.LastName {
			changes.Add("LastName", oldUser.LastName, newUser.LastName)
			oldUser.LastName = newUser.LastName
		}

		if oldUser.Nickname != newUser.Nickname {
			changes.Add("Nickname", oldUser.Nickname, newUser.Nickname)
			oldUser.Nickname = newUser.Nickname
		}

		if oldUser.Email != newUser.Email {
			changes.Add("Email", oldUser.Email, newUser.Email)
			oldUser.Email = newUser.Email
		}

		if oldUser.Country != newUser.Country {
			changes.Add("Country", oldUser.Country, newUser.Country)
			oldUser.Country = newUser.Country
		}
		// TODO: do we interested in change of updated_at value?
		// TODO: do we consider update without changes as an actual update?

		// TODO: send notification about update of the user
		return nil
	})

	return err
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.storage.WithoutTx(ctx, func(runner storage.Runner) error {
		return s.storage.Delete(ctx, runner, id)
	}); err != nil {
		return fmt.Errorf("delete user %q: %w", id, err)
	}

	// TODO: send notification about deletion of the user

	return nil
}

type Changes map[string]Change

type Change struct {
	Old, New interface{}
}

func (cs Changes) Add(property string, o, n interface{}) {
	cs[property] = Change{Old: o, New: n}
}
