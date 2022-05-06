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
	ErrEditConflict = errors.New("edit conflict")
)

type Models struct {
	Books interface {
		GetAll(filters Filters) ([]*Book, Metadata, error)
		GetBookSuggestions (typeSearch string, filters Filters) ([]*Book, error)
		AdvanceFilterBooks (filters Filters) ([]*Book, Metadata, error)
	}
}

func NewModel(db *sql.DB, es *elasticsearch.Client) Models {
	return Models{
		Books: BookModel{DB: db, ES: es},
	}
}