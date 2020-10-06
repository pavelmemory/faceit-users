package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type User struct {
	ID        string
	Version   int64 // TODO: optimistic locking to prevent unexpected concurrent modification by other instances
	FirstName string
	LastName  string
	Nickname  string
	Email     string
	Password  string
	Country   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (p *Postgres) Persist(ctx context.Context, run Runner, user User) (string, error) {
	// TODO: the simplest way to handle passwords management taken from official doc.
	// https://www.postgresql.org/docs/11/pgcrypto.html
	// Hashed in this way passwords are not easy portable to other storage.
	const query = `
		INSERT INTO users(first_name, last_name, nickname, email, country, password, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, CRYPT($6, GEN_SALT('md5')), $7, $8) 
		RETURNING id`

	var id string

	res := run.QuerySingle(ctx, query, user.FirstName, user.LastName, user.Nickname, user.Email, user.Country, user.Password, user.CreatedAt, user.UpdatedAt)
	if err := convertError(res.Scan(&id)); err != nil {
		return "", fmt.Errorf("query single: %w", err)
	}
	return id, nil
}

func (p *Postgres) Retrieve(ctx context.Context, run Runner, id string, forUpdate bool) (User, error) {
	var query = []string{`
		SELECT first_name, last_name, nickname, email, country, created_at, updated_at
		FROM users
		WHERE id = $1`,
	}

	if forUpdate {
		query = append(query, `FOR UPDATE`)
	}

	u := User{ID: id}

	res := run.QuerySingle(ctx, strings.Join(query, " "), id)
	if err := convertError(res.Scan(&u.FirstName, &u.LastName, &u.Nickname, &u.Email, &u.Country, &u.CreatedAt, &u.UpdatedAt)); err != nil {
		return User{}, fmt.Errorf("query single: %w", err)
	}

	return u, nil
}

func (p *Postgres) Update(ctx context.Context, run Runner, id string, user User) (User, error) {
	const query = `
		WITH old_state AS (
			SELECT first_name, last_name, nickname, email, country, created_at, updated_at 
			FROM users 
			WHERE id = $1
		)
		, update_state AS (
			UPDATE users 
			SET 
				first_name = $2, 
				last_name = $3, 
				nickname = $4, 
				email = $5, 
				country = $6, 
				updated_at = $7
			WHERE id = $1
		)
		SELECT * FROM old_state`

	u := User{ID: id}

	res := run.QuerySingle(ctx, query, id, user.FirstName, user.LastName, user.Nickname, user.Email, user.Country, user.UpdatedAt)
	if err := convertError(res.Scan(&u.FirstName, &u.LastName, &u.Nickname, &u.Email, &u.Country, &u.CreatedAt, &u.UpdatedAt)); err != nil {
		return User{}, fmt.Errorf("query single: %w", err)
	}

	return u, nil
}

func (p *Postgres) Delete(ctx context.Context, run Runner, id string) error {
	const query = `DELETE FROM users WHERE id = $1 RETURNING true`
	var confirmation sql.NullBool
	if err := run.QuerySingle(ctx, query, id).Scan(&confirmation); err != nil {
		return convertError(err)
	}
	return nil
}
