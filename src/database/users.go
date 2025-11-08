package database

import (
	"github.com/jmoiron/sqlx"
)

func GetUserByMaxID(db *sqlx.DB, userID int64) (*User, error) {
	user := new(User)
	err := db.Get(user, `SELECT * FROM users WHERE usermax_id = $1`, userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func GetUserRole(db *sqlx.DB, userID int64) (string, error) {
	var roleName string

	err := db.Get(&roleName, `SELECT r.role_name FROM users u JOIN roles r ON u.role_id = r.role_id  WHERE u.usermax_id = $1`, userID)
	if err != nil {
		return "", err
	}

	return roleName, nil
}
