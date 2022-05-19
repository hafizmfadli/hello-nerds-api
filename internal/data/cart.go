package data

import (
	"context"
	"database/sql"
	"time"
)

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

// Insert an new record in the database for the cart. Note that the id and created_at
// field are all automatically generated by our database.
func (m CartModel) Insert(cart *Cart) error {
	query := `
	INSERT INTO carts(user_id, updated_edited_id, quantity, total_price)
	VALUES(?, ?, ?, (SELECT ? * price total_price FROM updated_edited WHERE id = ?));`

	args := []interface{}{cart.UserID, cart.UpdatedEditedID, cart.Quantity, cart.Quantity, cart.UpdatedEditedID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	cart.ID = id
	cart.CreatedAt = time.Now()

	return nil
}
