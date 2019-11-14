// password-hashing connects to a psql database, inserts a row of data gathered from the user, and asks for the previously
// entered password to compare it to the hash that was stored in the database.
package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"os"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

type dbStuff struct {
	database  *sql.DB
	age       int
	firstName string
	lastName  string
	email     string
	password  string
	id        int
}

var dbInfo dbStuff

func main() {
	// Connect to DB
	dbConnect()
	defer dbInfo.database.Close()

	// Get user info
	getUserInfo()

	// Add to DB
	insertWithPassword(dbInfo.age, dbInfo.firstName, dbInfo.lastName,
		dbInfo.email, dbInfo.password)

	// Ask for password and check hash.
	dbPassword := []byte(searchLastInsert())
	password := []byte(getOnlyPassword())

	fmt.Printf("Plaintext password:\n dbPassword: %q\n, password %q\n\n", string(dbPassword), string(password))

	err := bcrypt.CompareHashAndPassword(dbPassword, password)
	if err != nil {
		fmt.Printf("Hashes do not match!\n\n Hash1: %q\n Hash2: %q\n", dbPassword, password)
	} else {
		fmt.Printf("Hashes match!\n\n Hash1: %q\n Hash2: %q\n", dbPassword, password)

	}

}

func dbConnect() {
	host := ""
	port := 5432
	user := ""
	password := ""
	dbname := ""

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	dbInfo.database = db

	fmt.Println("Database connection successful!")

}

func insertWithPassword(age int, first, last, email, password string) {
	sqlStatement := `
	INSERT INTO users (age, email, first_name, last_name, password)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id`

	err := dbInfo.database.QueryRow(sqlStatement, age, email, first, last, password).Scan(&dbInfo.id)
	if err != nil {
		panic(err)
	}
	fmt.Println("New record ID is: ", dbInfo.id)
}

func searchLastInsert() string {
	sqlStatement := `
	SELECT password FROM users
	WHERE id = $1`

	var password string
	err := dbInfo.database.QueryRow(sqlStatement, dbInfo.id).Scan(&password)
	if err != nil {
		panic(err)
	}
	return password
}

func getUserInfo() {
	reader := bufio.NewReader(os.Stdin)

	// Age
	fmt.Print("Enter Age: ")
	age, err := reader.ReadString('\n')
	dbInfo.age, err = strconv.Atoi(strings.TrimSuffix(age, "\n"))
	if err != nil {
		panic(err)
	}

	// Email
	fmt.Print("Enter Email: ")
	email, err := reader.ReadString('\n')
	dbInfo.email = strings.TrimSuffix(email, "\n")

	// First Name
	fmt.Print("Enter First Name: ")
	first, err := reader.ReadString('\n')
	dbInfo.firstName = strings.TrimSuffix(first, "\n")

	// Last Name
	fmt.Print("Enter Last Name: ")
	last, err := reader.ReadString('\n')
	dbInfo.lastName = strings.TrimSuffix(last, "\n")

	// Password
	fmt.Print("Enter Password: ")
	password, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	password = strings.TrimSuffix(password, "\n")
	dbInfo.password = hashPassword(password)

}

func getOnlyPassword() string {
	reader := bufio.NewReader(os.Stdin)

	// Password
	fmt.Print("Enter Password: ")
	password, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	password = strings.TrimSuffix(password, "\n")
	return password
}

func hashPassword(passwordString string) string {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(passwordString), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(passwordHash)
}
