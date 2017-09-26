package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bjwbell/renfish/auth"
	_ "github.com/mattn/go-sqlite3"
)

const DbName = "renfishdb"

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

func Exists(name string) bool {
	success, _ := exists("./" + name + ".sqlite")
	return success
}

func Create(name string) bool {
	if Exists(name) {
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
Name text, SubDomain text, ContainerId text);
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

func SaveSite(userEmail, siteName, ip string) (int64, bool) {
	userName := ""
	return DbInsert(DbName, userEmail, userName, siteName, ip)
}

func DbGetSiteNames() []string {
	siteNames := []string{}
	if !Exists(DbName) {
		return []string{}
	}
	db, err := sql.Open("sqlite3", "./"+DbName+".sqlite")
	if err != nil {
		auth.LogError("Couldn't open database (" + DbName + ")")
		log.Fatal(err)
		return siteNames
	}
	defer db.Close()
	rows, err := db.Query(`select
                               id,
                               subdomain from users`)
	if err != nil {
		auth.LogError("Couldn't query database (" + DbName + ")")
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var siteName string
		rows.Scan(&id, &siteName)
		siteNames = append(siteNames, siteName)
	}
	return siteNames
}

func DbInsert(dbName, userEmail, name, subdomain, containerId string) (int64, bool) {

	if !Exists(dbName) {
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
	stmt, err := tx.Prepare("insert into users(id, timestamp, email, name, subdomain, containerId) values(?, ?, ?, ?, ?, ?)")
	if err != nil {
		auth.LogError("Couldn't prepare insert in database (" + dbName + ")" +
			", userEmail (" + userEmail + ")")
		log.Fatal(err)
		return -1, false
	}
	defer stmt.Close()
	var timestamp = time.Now()
	result, err := stmt.Exec(nil, timestamp, userEmail, name, subdomain, containerId)
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

func DbUpdate(dbName string, userEmail, subdomain, containerId string) bool {

	if !Exists(dbName) {
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
	stmt, err := tx.Prepare("insert into users(id, timestamp, useremail, subdomain, containerId) values(?, ?, ?, ?, ?)")
	if err != nil {
		auth.LogError("Couldn't prepare update insert in database (" + dbName + ")" +
			", userEmail (" + userEmail + ")")
		log.Fatal(err)
		return false
	}
	defer stmt.Close()
	var timestamp = time.Now().Format("2006-01-02 15:04:05.000000000")
	_, err = stmt.Exec(nil, timestamp, userEmail, subdomain, containerId)
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
