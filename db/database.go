package db

import (
	"event"
	"fmt"
	"log"
	"os"
	"user"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func Configure() {
	dataSource := fmt.Sprintf("%s:%s@tcp(%s)/%s", os.Getenv("MYSQL_USER"), os.Getenv("MYSQL_PASSWORD"), os.Getenv("MYSQL_URL"), os.Getenv("MYSQL_DATABASE"))

	DB, err := sqlx.Connect("mysql", dataSource)
	if err != nil {
		log.Fatalln(err)
	}

	pingErr := DB.Ping()

	if pingErr != nil {
		log.Fatal(pingErr)
	}

	fmt.Println("Connected to MySQL Database")

	user.DB = DB
	event.DB = DB
}
