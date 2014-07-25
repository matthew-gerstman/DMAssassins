package main

import (
	"errors"
	"github.com/gorilla/sessions"
	"net/http"
)

//yes i know i need a real secret key and i should read it from a config file
var store = sessions.NewCookieStore([]byte("some-thing-very-secret"))

// Create a session this will probably be rewritten later with basic auth
func postSession(w http.ResponseWriter, r *http.Request) (interface{}, *ApplicationError) {
	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")
	user, err := GetUserByUsername(username)

	if err != nil {
		return nil, err
	}
	bytePW := []byte(password)
	valid := user.CheckPassword(bytePW)

	if valid {
		session, _ := store.Get(r, "DMAssassins")
		session.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   1800,
			HttpOnly: true,
		}
		session.Values["user_id"] = user.User_id
		session.Save(r, w)
	}

	return valid, nil
}

// Kill a session this will probably be rewritten later with basic auth
func deleteSession(w http.ResponseWriter, r *http.Request) (interface{}, *ApplicationError) {
	session, _ := store.Get(r, "DMAssassins")
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}

	return session.Save(r, w), nil
}

func SessionHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var obj interface{}
		var err *ApplicationError

		switch r.Method {
		case "POST":
			obj, err = postSession(w, r)
		case "DELETE":
			obj, err = deleteSession(w, r)
		default:
			obj = nil
			msg := "Not Found"
			err := errors.New("Invalid Http Method")
			err = NewApplicationError(msg, err, ErrCodeInvalidMethod)
		}
		WriteObjToPayload(w, r, obj, err)
	}
}
