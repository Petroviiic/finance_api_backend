package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type UserStorage struct {
	db *sql.DB
}

type User struct {
	ID         int64    `json:"id"`
	FirstName  string   `json:"first_name"`
	LastName   string   `json:"last_name"`
	Username   string   `json:"username"`
	Email      string   `json:"email"`
	Password   password `json:"-"`
	Created_at string   `json:"created_at"`
	Role       string   `json:"role"`
	IsActive   bool     `json:"is_active"`
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
	query := `	
			SELECT 
				id, 
				email, 
				first_name, 
				last_name, 
				username, 
				password, 
				created_at, 
				role, 
				is_active 
			FROM users 
			WHERE username = $1
		`

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
		&user.FirstName,
		&user.LastName,
		&user.Username,
		&user.Password.Hash,
		&user.Created_at,
		&user.Role,
		&user.IsActive,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UserStorage) GetById(ctx context.Context, userId int64) (*User, error) {
	query := `	
			SELECT 
				id, 
				email, 
				first_name, 
				last_name, 
				username, 
				password, 
				created_at, 
				role, 
				is_active 
			FROM users 
			WHERE id = $1
		`

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
		&user.FirstName,
		&user.LastName,
		&user.Username,
		&user.Password.Hash,
		&user.Created_at,
		&user.Role,
		&user.IsActive,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UserStorage) RegisterUser(ctx context.Context, user *User) (int64, error) {
	query := `
			INSERT INTO users 
			(email, 
			username, 
			first_name, 
			last_name, 
			password,
			role,
			is_active) 
			VALUES 
			($1, $2, $3, $4, $5, $6, $7) RETURNING id;
		`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var userId int64
	err := u.db.QueryRowContext(
		ctx,
		query,
		user.Email,
		user.Username,
		user.FirstName,
		user.LastName,
		user.Password.Hash,
		user.Role,
		user.IsActive,
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

func (u *UserStorage) UpdateUserStatus(ctx context.Context, id int64, isActive bool) error {
	query := `UPDATE users SET is_active = $1 WHERE id = $2`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := u.db.ExecContext(ctx, query, isActive, id)

	if err != nil {
		return err
	}
	return nil
}

func (u *UserStorage) UpdateUserRole(ctx context.Context, id int64, role string) error {
	query := `UPDATE users SET role = $1 WHERE id = $2`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := u.db.ExecContext(ctx, query, role, id)

	if err != nil {
		return err
	}
	return nil
}

func (u *UserStorage) GetAllUsers(ctx context.Context) ([]*User, error) {
	query := `
        SELECT id, email, username, first_name, last_name, role, is_active, created_at 
        FROM users
    `

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := u.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Username,
			&user.FirstName,
			&user.LastName,
			&user.Role,
			&user.IsActive,
			&user.Created_at,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (u *UserStorage) DeleteUser(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := u.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}
