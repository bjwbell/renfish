package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bjwbell/renfish/auth"
	_ "github.com/mattn/go-sqlite3"
)

const ActionInsert = "insert"
const ActionUpdate = "update"
const ActionUndoUpdate = "undoupdate"
const ActionRemove = "remove"
const ActionUndoRemove = "undoremove"

// exists returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func dbExists(name string) bool {
	success, _ := exists("./" + name + ".sqlite")
	if !success {
		auth.LogError("dbExists: database (" + name + ") doesnt exist")
	}
	return success
}

func dbCreate(name string) bool {
	if dbExists(name) {
		auth.LogError("Database (" + name + ") already exists, RECREATING")
		if err := os.Remove("./" + name + ".sqlite"); err != nil {
			auth.LogError("Error removing db:" + fmt.Sprintf("%v", err))
		}
	}
	db, err := sql.Open("sqlite3", "./"+name+".sqlite")
	if err != nil {
		auth.LogError(fmt.Sprintf("Couldn't create database ("+
			name+"), ERROR: %v", err))
		log.Fatal(err)
		return false
	}
	defer db.Close()
	sqlStmt := `
	create table users
(id integer not null primary key,
TimeStamp text, Email text,
Name text, Subdomain text, dockerimageid text);
	delete from users;
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		auth.LogError(fmt.Sprintf("Couldn't create table, database ("+
			name+"), ERROR (%q: %s\n)", err, sqlStmt))
		return false
	}
	return true
}

func dbInsert(dbName, userEmail, name, subdomain string) (int64, bool) {

	if !dbExists(dbName) {
		return -1, false
	}
	db, err := sql.Open("sqlite3", "./"+dbName+".sqlite")
	if err != nil {
		auth.LogError("Couldn't open database (" + dbName + ")" +
			", userEmail (" + userEmail + ")")
		log.Fatal(err)
		return -1, false
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		auth.LogError("Couldn't exec begin for database (" + dbName + ")" +
			", userEmail (" + userEmail + ")")
		log.Fatal(err)
		return -1, false
	}
	stmt, err := tx.Prepare("insert into users(id, timestamp, email, name, subdomain) values(?, ?, ?, ?, ?)")
	if err != nil {
		auth.LogError("Couldn't prepare insert in database (" + dbName + ")" +
			", userEmail (" + userEmail + ")")
		log.Fatal(err)
		return -1, false
	}
	defer stmt.Close()
	var timestamp = time.Now()
	result, err := stmt.Exec(nil, timestamp, userEmail, name, subdomain)
	if err != nil {
		auth.LogError("Couldn't exec insert in database (" + dbName + ")" +
			", userEmail (" + userEmail + ")")
		log.Fatal(err)
		return -1, false
	}
	if err = tx.Commit(); err != nil {
		auth.LogError("Couldn't exec insert in database (" + dbName + ")" +
			", userEmail (" + userEmail + ")")
		log.Fatal(err)
		return -1, false
	}
	id, _ := result.LastInsertId()
	return id, true
}

func dbUpdate(dbName string, userEmail, subdomain, dockerImageID string) bool {

	if !dbExists(dbName) {
		return false
	}
	fmt.Println("dbUpdate - userEmail:")
	fmt.Println("userEmail: ", userEmail, ", subdomain: ", subdomain)
	db, err := sql.Open("sqlite3", "./"+dbName+".sqlite")
	if err != nil {
		auth.LogError("Couldn't open database (" + dbName + ")" +
			", userEmail (" + userEmail + ")")
		log.Fatal(err)
		return false
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		auth.LogError("Couldn't exec begin for database (" + dbName + ")" +
			", userEmail (" + userEmail + ")")
		log.Fatal(err)
		return false
	}
	stmt, err := tx.Prepare("insert into users(id, timestamp, useremail, subdomain, dockerimageid) values(?, ?, ?, ?, ?)")
	if err != nil {
		auth.LogError("Couldn't prepare update insert in database (" + dbName + ")" +
			", userEmail (" + userEmail + ")")
		log.Fatal(err)
		return false
	}
	defer stmt.Close()
	var timestamp = time.Now().Format("2006-01-02 15:04:05.000000000")
	_, err = stmt.Exec(nil, timestamp, userEmail, subdomain)
	if err != nil {
		auth.LogError("Couldn't exec update insert in database (" + dbName + ")" +
			", userEmail (" + userEmail + ")")
		log.Fatal(err)
		return false
	}
	if err = tx.Commit(); err != nil {
		auth.LogError("Couldn't exec update insert in database (" + dbName + ")" +
			", userEmail (" + userEmail + ")")
		log.Fatal(err)
		return false
	}
	return true
}
