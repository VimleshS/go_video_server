package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("sqlite3", "./videos.db")
	if err != nil {
		panic(err.Error())
	}
}

func createDB() {
	db.Exec(`Drop table Videos;
		     Drop table Admin;`)

	sqlStmt := `CREATE TABLE Videos (
			Id          INTEGER       PRIMARY KEY,
			Name        VARCHAR (100),
			IsDirectory BOOL,
			Path 		VARCHAR (250),
			Description TEXT
			);
			CREATE INDEX video_name_idx ON Videos (Name);
			CREATE TABLE Admin (
			Email       STRING (40) PRIMARY KEY
			);`

	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	qStmt, err := tx.Prepare("INSERT INTO Videos(Id, Name, IsDirectory, Path, Description) VALUES (?,?,?,?,?)")
	if err != nil {
		log.Fatal(err)
	}

	files := videoListHandler{}.getVideoFilesInDirectory()
	for _, file := range files {
		_, err1 := qStmt.Exec(file.ID, file.Name, file.IsDirectory, file.Path, "")
		if err1 != nil {
			log.Fatal(err)
		}
	}
	tx.Commit()
	fmt.Println("Synced")
}

func readvideoInfo() []FileInfo {
	dbFileInfo := []FileInfo{}

	rows, err := db.Query("select Id, Name, IsDirectory, Path, Description from Videos")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		var isDirectory bool
		var path string
		var desc string
		err = rows.Scan(&id, &name, &isDirectory, &path, &desc)
		if err != nil {
			log.Fatal(err)
		}
		dbFileInfo = append(dbFileInfo, FileInfo{ID: id, Name: name, IsDirectory: isDirectory, Path: path, Desc: desc})
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return dbFileInfo
}

func updateRec(id int, desc string) error {
	stmt, err := db.Prepare("UPDATE Videos SET Description = ? WHERE Id = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(desc, id)
	if err != nil {
		return err
	}
	return nil
}

func isAdmin(email string) bool {
	rows, err := db.Query("Select email from Admin where email = ?", email)
	if err != nil {
		log.Fatal(err)
		return false
	}
	defer rows.Close()
	var admin string
	for rows.Next() {
		err = rows.Scan(&admin)
		if err != nil {
			log.Fatal(err)
		}
	}
	return (admin == email)
}
