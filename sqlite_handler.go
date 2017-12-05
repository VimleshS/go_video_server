package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func init() {
	db, err := sql.Open("sqlite3", "./videos.db")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	stmt, err := db.Prepare(`CREATE TABLE Videos (
		Id          INTEGER       PRIMARY KEY,						
		Name        VARCHAR (100),
		IsDirectory BOOL,
		Description TEXT
		);`)

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	res, err := stmt.Exec()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(res.RowsAffected)

	files := videoListHandler{}.getVideoFilesInDirectory()
	for _, file := range files {
		stmt, err := db.Prepare("INSERT INTO Videos(Id, Name, IsDirectory, Description) VALUES (?,?,?,?)")
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		_, err1 := stmt.Exec(file.ID, file.Name, file.IsDirectory, file.Path)
		if err1 != nil {
			fmt.Println(err.Error())
			return
		}
	}
	fmt.Println("Sync complete")
}
