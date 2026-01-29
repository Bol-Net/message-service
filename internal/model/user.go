package model

import (
	"messaging-service/internal/db"
	"strconv"
	
	"github.com/lib/pq"
)

type User struct {
	ID    int    `db:"id" json:"id"`
	Name  string `db:"name" json:"name"`
	Email string `db:"email" json:"email"`
	Role  int    `db:"role" json:"role"`
}

// GetUserByID fetches a user by their ID
func GetUserByID(userID string) (*User, error) {
	query := `SELECT id, name, email, role FROM users WHERE id = $1`
	
	var user User
	err := db.DB.QueryRow(query, userID).Scan(&user.ID, &user.Name, &user.Email, &user.Role)
	if err != nil {
		return nil, err
	}
	
	return &user, nil
}

// GetAllUsers fetches all users (for online users list)
func GetAllUsers() ([]User, error) {
	query := `SELECT id, name, email, role FROM users ORDER BY name`
	
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Role); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	
	return users, nil
}

// GetUsersByIDs fetches multiple users by their IDs
func GetUsersByIDs(userIDs []string) (map[string]User, error) {
	if len(userIDs) == 0 {
		return make(map[string]User), nil
	}
	
	query := `SELECT id, name, email, role FROM users WHERE id = ANY($1)`
	
	// Use pq.Array for PostgreSQL array parameter
	rows, err := db.DB.Query(query, pq.Array(userIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	userMap := make(map[string]User)
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Role); err != nil {
			return nil, err
		}
		userMap[strconv.Itoa(user.ID)] = user
	}
	
	return userMap, nil
}
