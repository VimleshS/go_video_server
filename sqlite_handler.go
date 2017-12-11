package main

import (
	"net/http"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

func createDb(w http.ResponseWriter, r *http.Request) {
	createDB()
	w.Write([]byte("DB Recreated with files in Directory `static/videos`"))
}

func edit(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.FormValue("Id"))
	desc := r.FormValue("description")

	session, _ := store.Get(r, "user-details")
	_user := session.Values["user"]
	sessionUser := ""
	if _user != nil {
		sessionUser = _user.(*User).Email
	} else {
		http.Error(w, "User Not found.", http.StatusBadRequest)
	}

	err := updateRec(id, sessionUser, desc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
