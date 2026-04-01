package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type UserStorage struct {
	db *sql.DB
}

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Password  password  `json:"-"`
	Points    int       `json:"points"`
	CreatedAt time.Time `json:"created_at"`
}

type password struct {
	Plain string
	Hash  []byte
}

func (p *password) Set(plain string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), 14)
	if err != nil {
		return err
	}

	p.Plain = plain
	p.Hash = hash
	return nil
}

func (p *password) ValidatePassword(plain string) bool {
	if err := bcrypt.CompareHashAndPassword(p.Hash, []byte(plain)); err != nil {
		return false
	}
	return true
}
func (u *UserStorage) GetByUsername(ctx context.Context, username string) (*User, error) {
	query := `	SELECT id, email, username, password, points, created_at FROM users 
				WHERE username = $1`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user := &User{}
	err := u.db.QueryRowContext(
		ctx,
		query,
		username,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Password.Hash,
		&user.Points,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UserStorage) GetById(ctx context.Context, userId int64) (*User, error) {
	query := `	SELECT id, email, username, password, points, created_at FROM users 
				WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user := &User{}
	err := u.db.QueryRowContext(
		ctx,
		query,
		userId,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Password.Hash,
		&user.Points,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UserStorage) RegisterUser(ctx context.Context, user *User) (int64, error) {
	query := `
			INSERT INTO users (email, username, password) VALUES ($1, $2, $3) RETURNING id;
		`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var userId int64
	err := u.db.QueryRowContext(
		ctx,
		query,
		user.Email,
		user.Username,
		user.Password.Hash,
	).Scan(
		&userId,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				return -1, ERROR_DUPLICATE_KEY_VALUE
			}
		}
		return -1, err
	}
	return userId, nil
}
