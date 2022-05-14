package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Define a Permissions slice, which we will use to hold the permission codes
// (like "books:read" and "books:write") for a single user.
type Permissions []string

// Add a helper method to check whether the permissions slice contains a specific
// permission code.
func (p Permissions) Include(code string) bool {
	for i := range p {
		if code == p[i] {
			return true
		}
	}
	return false
}

// Define the PermissionModel type.
type PermissionModel struct {
	DB *sql.DB
}

// The GetAllForUser() method returns all permission codes for a specific user in a
// Permissions slice.
func (m PermissionModel) GetAllForUser(userID int64) (Permissions, error) {
	query := `
		SELECT permissions.code
		FROM permissions
		INNER JOIN users_permissions ON users_permissions.permission_id = permissions.id
		INNER JOIN users ON users_permissions.user_id = users.id
		WHERE users.id = ?`
	
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions Permissions

	for rows.Next() {
		var permission string

		err := rows.Scan(&permission)
		if err != nil {
			return nil, err
		}

		permissions = append(permissions, permission)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

// Add the provided permission codes for a specific user. notice that we're using a
// variadic parameter for the codes so that we can assign multiple permissions in a
// single call
func (m PermissionModel) AddForUser(userID int64, codes ...string) error {
	
	var args string
	codesLen := len(codes)
	for i := 0; i < codesLen; i++ {
		args += fmt.Sprintf(`"%s"`, codes[i]) 
		if i != (codesLen - 1) {
			args += ","
		}
	}

	query := fmt.Sprintf(`
	INSERT INTO users_permissions
	SELECT ?, permissions.id FROM permissions WHERE permissions.code IN (%s)`, args)

	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, userID)
	return err
}