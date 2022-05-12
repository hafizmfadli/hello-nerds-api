package data

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/hafizmfadli/hello-nerds-api/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

// custom ErrDuplicateEmail error
var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
}

type password struct {
	plainText *string
	hash []byte
}

// The Set() method calculates the bcrypt hash of a plaintext password, and
// stores both the hash and the plaintext version in the struct
func (p *password) Set(plaintextPassword string) error {
	_, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}
	p.plainText = &plaintextPassword
	// p.hash = hash

	return nil
}

// Matches() method checks whether the provided plaintext password matches
// the hashed password stored in the struct, returning true if match and false
// otherwise
func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func ValidateEmail (v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid ")
}

func ValidatePasswordPlaintext (v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateConfirmPassword (v *validator.Validator, password, confirmPassword string) {
	v.Check(password == confirmPassword, "confirm_password", "must be same as password")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.FirstName != "", "first_name", "must be provided")
	v.Check(len(user.FirstName) >= 3, "first_name", "must be at least 3 characters long")
	v.Check(len(user.FirstName) <= 30, "first_name", "must not be more than 30 characters long")

	v.Check(user.LastName != "", "last_name", "must be provided")
	v.Check(len(user.LastName) >= 3, "last_name", "must be at least 3 characters long")
	v.Check(len(user.LastName) <= 30, "last_name", "must not be more than 30 characters long")

	ValidateEmail(v, user.Email)

	if user.Password.plainText != nil {
		ValidatePasswordPlaintext(v, *user.Password.plainText)
	}

	// If the password hash is ever nil, this will be due to a logic error in our
	// codebase (probably because we forgot to set a password for the user). It's a
	// useful sanity check to include here, but it's not a problem with the data
	// provided by the client. So rather than adding an error to the validation map we
	// raise a panic instead.
	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

type UserModel struct {
	DB *sql.DB
}

// Insert an new record in the database for the user. Note that the id, created_at, and activated
// field are all automatically generated by our database.
func (m UserModel) Insert(user *User) error {
	query := `
		INSERT INTO users (first_name, last_name, email, password_hash)
		VALUES (?, ?, ?, ?)`
	
	args := []interface{}{user.FirstName, user.LastName, user.Email, user.Password.hash}

	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == 1062 && strings.Contains(mysqlErr.Message, "users.email"){
				return ErrDuplicateEmail
			}
		}else {
			return err
		}
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = id
	user.CreatedAt = time.Now()
	user.Activated = false

	return nil
}

// GetByEmail retrieve the user details from the database based on the user's email address.
func (m UserModel) GetByEmail(email string) (*User, error) {
	query := `
		SELECT id, created_at, first_name, last_name, email, password_hash, activated
		FROM users
		WHERE email = ?`
	
	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

// Update the details for a specific user.
func (m UserModel) Update (user *User) error {
	query := `
		UPDATE users
		SET first_name = ?, last_name = ?, email = ?, password_hash = ?, activated = ?
		WHERE id = ?`
	
	args := []interface{}{
		user.FirstName,
		user.LastName,
		user.Email,
		user.Password.hash,
		user.Activated,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == 1062 && strings.Contains(mysqlErr.Message, "users.email"){
				return ErrDuplicateEmail
			}
		}else {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				return ErrEditConflict
			default:
				return err
			}
		}
	}

	return nil
}