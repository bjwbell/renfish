package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bjwbell/renfish/auth"
	_ "github.com/mattn/go-sqlite3"
)

const DbName = "renfish.db"

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
Name text, Subdomain text, IP text, dockerimageid text);
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

func DbInsert(dbName, userEmail, name, subdomain, ip string) (int64, bool) {

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
	stmt, err := tx.Prepare("insert into users(id, timestamp, email, name, subdomain, ip) values(?, ?, ?, ?, ?, ?)")
	if err != nil {
		auth.LogError("Couldn't prepare insert in database (" + dbName + ")" +
			", userEmail (" + userEmail + ")")
		log.Fatal(err)
		return -1, false
	}
	defer stmt.Close()
	var timestamp = time.Now()
	result, err := stmt.Exec(nil, timestamp, userEmail, name, subdomain, ip)
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

func DbUpdate(dbName string, userEmail, subdomain, ip, dockerImageID string) bool {

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
	stmt, err := tx.Prepare("insert into users(id, timestamp, useremail, subdomain, ip, dockerimageid) values(?, ?, ?, ?, ?, ?)")
	if err != nil {
		auth.LogError("Couldn't prepare update insert in database (" + dbName + ")" +
			", userEmail (" + userEmail + ")")
		log.Fatal(err)
		return false
	}
	defer stmt.Close()
	var timestamp = time.Now().Format("2006-01-02 15:04:05.000000000")
	_, err = stmt.Exec(nil, timestamp, userEmail, subdomain, ip, dockerImageID)
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

func DbGetIPs(dbName string) []string {
	if !Exists(DbName) {
		auth.LogError("DbGetIPs: Database doesn't exist (" + dbName + ")")
		Create(dbName)
	}
	db, err := sql.Open("sqlite3", "./"+dbName+".sqlite")
	if err != nil {
		auth.LogError("Couldn't read database (" + dbName + ")")
		log.Fatal(err)
	}
	defer db.Close()
	rows, err := db.Query(`select
                               id,
                               ip from users where ip is not null`)
	if err != nil {
		auth.LogError("Couldn't query database (" + dbName + ")")
		log.Fatal(err)
	}
	defer rows.Close()
	ips := []string{}
	var id int
	var ip string
	for rows.Next() {
		if err := rows.Scan(&id, &ip); err != nil {
			auth.LogError("Couldn't query database (" + dbName + ")")
			log.Fatal(err)
		}
		ips = append(ips, ip)
	}
	return ips
}

func IPToHex(ip string) uint32 {
	parts := strings.Split(ip, ".")
	part0, err := strconv.Atoi(parts[0])
	if err != nil {
		auth.LogError("Couldn't parse ip address to int")
		log.Fatal(err)
	}
	part1, err := strconv.Atoi(parts[1])
	if err != nil {
		auth.LogError("Couldn't parse ip address to int")
		log.Fatal(err)
	}
	part2, err := strconv.Atoi(parts[2])
	if err != nil {
		auth.LogError("Couldn't parse ip address to int")
		log.Fatal(err)
	}
	part3, err := strconv.Atoi(parts[3])
	if err != nil {
		auth.LogError("Couldn't parse ip address to int")
		log.Fatal(err)
	}
	return uint32(part0)<<24 + uint32(part1)<<16 + uint32(part2)<<8 + uint32(part3)
}

func HexToIP(ip uint32) string {
	part0 := strconv.Itoa(int(ip >> 24))
	part1 := strconv.Itoa(int((ip << 8) >> 24))
	part2 := strconv.Itoa(int((ip << 16) >> 24))
	part3 := strconv.Itoa(int((ip << 24) >> 24))
	return part0 + "." + part1 + "." + part2 + "." + part3
}

const FirstIP = "172.19.0.2"

func DbGetNextAvailableIP(ips []string) string {
	var maxIP uint32
	for _, ip := range ips {
		if IPToHex(ip) > maxIP {
			maxIP = IPToHex(ip)
		}
	}
	if maxIP == 0 {
		return FirstIP
	} else {
		maxIP++
		return HexToIP(maxIP)
	}
}
