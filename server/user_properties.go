package main

import (
	"database/sql"
	"errors"
	"github.com/getsentry/raven-go"
	"strings"
)

func (user *User) SetUserProperty(key string, value string) (*User, *ApplicationError) {
	res, err := db.Exec(`UPDATE dm_user_properties SET value = $1 WHERE user_id = $2 AND key ILIKE $3`, value, user.User_id, key)
	if err != nil {
		return nil, NewApplicationError("Internal Error", err, ErrCodeDatabase)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return nil, NewApplicationError("Internal Error", err, ErrCodeDatabase)
	}

	if rowsAffected == 0 {
		res, err := db.Exec(`INSERT INTO dm_user_properties (user_id, key, value) VALUES ($1,$2,$3)`, user.User_id, key, value)
		if err != nil {
			return nil, NewApplicationError("Internal Error", err, ErrCodeDatabase)
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return nil, NewApplicationError("Internal Error", err, ErrCodeDatabase)
		}
		if rowsAffected == 0 {
			err = errors.New("Failed insert for " + key + " : " + value)
			return nil, NewApplicationError("Internal Error", err, ErrCodeDatabase)
		}
	}

	//_, err := db.Exec(`INSERT INTO dm_user_properties (user_id, key, value) VALUES ($1, $2, $3)`, user.User_id, key, value)
	user.Properties[key] = value
	return user, nil
}

func (user *User) GetUserProperty(key string) (string, *ApplicationError) {
	var property string
	err := db.QueryRow(`SELECT value FROM dm_user_properties WHERE user_id = $1 AND key ILIKE $2`, user.User_id, key).Scan(&property)
	if err != nil {
		return "", NewApplicationError("Internal Error", err, ErrCodeDatabase)
	}
	return property, nil
}

func (user *User) GetUserProperties() (map[string]string, *ApplicationError) {

	properties := make(map[string]string)

	rows, err := db.Query(`SELECT key, value FROM dm_user_properties WHERE user_id = $1`, user.User_id)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, NewApplicationError("Internal Error", err, ErrCodeDatabase)
	}
	for rows.Next() {
		var key, value string

		err := rows.Scan(&key, &value)
		if err == nil {
			key = strings.ToLower(key)
			properties[key] = value
		} else {
			appErr := NewApplicationError("Error getting user properties", err, ErrCodeDatabase)
			LogWithSentry(appErr, map[string]string{"user_id": user.User_id}, raven.WARNING)
		}

	}
	return properties, nil
}