package storage

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ERROR_NO_ROWS_AFFECTED        = errors.New("no rows affected")
	ERROR_DUPLICATE_KEY_VALUE     = errors.New("record already exists")
	ERROR_NOT_ENOUGH_POINTS_FUNDS = errors.New("not enough funds")
	ERROR_ALREADY_OWN_MESSAGE     = errors.New("you already own the message")
)

type SQLCommon interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type Storage struct {
	UserStorage interface {
		GetById(context.Context, int64) (*User, error)
		GetByUsername(ctx context.Context, username string) (*User, error)
		RegisterUser(ctx context.Context, user *User) (int64, error)
		UpdateUserRole(ctx context.Context, id int64, role string) error
		UpdateUserStatus(ctx context.Context, id int64, isActive bool) error
		GetAllUsers(ctx context.Context) ([]*User, error)
		DeleteUser(ctx context.Context, id int64) error
	}
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{
		UserStorage: &UserStorage{db},
	}
}

func NewTx(ctx context.Context, db *sql.DB, function func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)

	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if err := function(tx); err != nil {
		return err
	}

	return tx.Commit()
}
