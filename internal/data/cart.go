package data

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

// custom ErrDuplicateUserAndBook error
var (
	ErrDuplicateUserAndBook = errors.New("duplicate user and book")
)

// UpdatedEditedID is id of book
type Cart struct {
	ID              int64     `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	Quantity        int64     `json:"quantity"`
	TotalPrice      int64     `json:"total_price"`
	UserID          int64     `json:"user_id"`
	UpdatedEditedID int64     `json:"updated_edited_id"`
}

type CartModel struct {
	DB *sql.DB
}

type BookDetail struct {
	ID         int64   `json:"id,omitempty"`
	ImgUrl     *string `json:"imgUrl,omitempty"`
	Title      *string `json:"title,omitempty"`
	Author     *string `json:"author,omitempty"`
	Identifier *string `json:"identifier,omitempty"`
	Price      int64   `json:"price,omitempty"`
	Stock      int     `json:"stock,omitempty"`
}

type CartDetail struct {
	BookDetail BookDetail `json:"book,omitempty"`
	Quantity   int64      `json:"quantity,omitempty"`
}

// Insert an new record in the database for the cart. Note that the id and created_at
// field are all automatically generated by our database.
func (m CartModel) Insert(cart *Cart) error {
	query := `
	INSERT INTO carts(user_id, updated_edited_id, quantity, total_price)
	SELECT ?, ?, ?, (SELECT ? * price total_price FROM updated_edited WHERE id = ?)
	WHERE (SELECT quantity FROM updated_edited WHERE id = ?) >= ?;`

	args := []interface{}{
		cart.UserID,
		cart.UpdatedEditedID,
		cart.Quantity,
		cart.Quantity,
		cart.UpdatedEditedID,
		cart.UpdatedEditedID,
		cart.Quantity,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == 1062 && strings.Contains(mysqlErr.Message, "carts.user_updated_edited_unique") {
				return ErrDuplicateUserAndBook
			}
		} else {
			return err
		}
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	cart.ID = id
	cart.CreatedAt = time.Now()

	return nil
}

// Update the quantity for a specific cart.
func (m CartModel) UpdateQuantity(cart *Cart) error {

	query := `
	UPDATE 
		carts
	INNER JOIN 
		updated_edited ON carts.updated_edited_id = updated_edited.id 
	SET 
		carts.quantity = ?, 
		carts.total_price = ? * updated_edited.price
	WHERE 
		carts.user_id = ? AND carts.updated_edited_id = ? AND updated_edited.quantity >= ?;`

	args := []interface{}{
		cart.Quantity,
		cart.Quantity,
		cart.UserID,
		cart.UpdatedEditedID,
		cart.Quantity,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

// GetByUserID retrieve the cart details from the database based on the user's ID.
func (m CartModel) GetByUserID(userID int64) (*[]CartDetail, error) {

	if userID < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
	SELECT
	carts.quantity, 
	updated_edited.id, 
	updated_edited.Coverurl, 
	updated_edited.Title, 
	updated_edited.Author, 
	updated_edited.Identifier, 
	updated_edited.price,
	updated_edited.quantity
		FROM carts
		INNER JOIN updated_edited ON carts.updated_edited_id = updated_edited.id
		WHERE carts.user_id = ?;`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	results, err := m.DB.QueryContext(ctx, query, userID)

	var cartDetails []CartDetail

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	for results.Next() {
		var cart Cart
		var book Book
		results.Scan(
			&cart.Quantity,
			&book.ID,
			&book.CoverUrl,
			&book.Title,
			&book.Author,
			&book.Identifier,
			&book.Price,
			&book.Quantity,
		)

		var bookDetail = BookDetail{
			ID:         book.ID,
			ImgUrl:     book.CoverUrl,
			Title:      book.Title,
			Author:     book.Author,
			Identifier: book.Identifier,
			Price:      book.Price,
			Stock:      book.Quantity,
		}

		var cartDetail = CartDetail{
			BookDetail: bookDetail,
			Quantity:   cart.Quantity,
		}

		cartDetails = append(cartDetails, cartDetail)
	}

	return &cartDetails, nil
}
