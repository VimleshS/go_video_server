package main

import (
	"fmt"
	"net/http"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

func createDb(w http.ResponseWriter, r *http.Request) {
	createDB()
	w.Write([]byte("DB Recreated with files in Directory `static/videos`"))
}

func edit(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.FormValue("Id"))
	fmt.Println(r.FormValue("description"))
	id, _ := strconv.Atoi(r.FormValue("Id"))
	desc := r.FormValue("description")
	err := updateRec(id, desc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
