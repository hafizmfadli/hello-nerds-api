package data

import (
	"database/sql"
	"errors"

	"github.com/elastic/go-elasticsearch/v7"
)

// Define a custom ErrRecordNotFound error. We'll return this from our Get() method when
// looking up a movie that doesn't exist in our database.
var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
	ErrNotEnoughStock = errors.New("not enough stock")
)

type Models struct {
	Books interface {
		GetAll(filters Filters) ([]*Book, Metadata, error)
		GetBookSuggestions(typeSearch string, filters Filters) ([]*Book, error)
		AdvanceFilterBooks(filters Filters) ([]*Book, Metadata, error)
		GetBook(id int64) (*Book, error)
	}
	Users       UserModel
	Tokens      TokenModel
	Permissions PermissionModel
	Indonesia IndonesiaModel
	Carts       CartModel
}

func NewModel(db *sql.DB, es *elasticsearch.Client) Models {
	return Models{
		Books:       BookModel{DB: db, ES: es},
		Users:       UserModel{DB: db},
		Tokens:      TokenModel{DB: db},
		Permissions: PermissionModel{DB: db},
		Indonesia: IndonesiaModel{DB: db},
		Carts:       CartModel{DB: db},
	}
}
