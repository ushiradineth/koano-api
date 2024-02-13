package mysql

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var Db *sql.DB

func Configure() {
	dataSource := fmt.Sprintf("%s:%s@tcp(%s)/%s", os.Getenv("MYSQL_USER"), os.Getenv("MYSQL_PASSWORD"), os.Getenv("MYSQL_URL"), os.Getenv("MYSQL_DATABASE"))

	Db, err := sqlx.Connect("mysql", dataSource)
	if err != nil {
		log.Fatalln(err)
	}

	pingErr := Db.Ping()

	if pingErr != nil {
		log.Fatal(pingErr)
	}

	fmt.Println("Connected to MySQL Database")

	Db.MustExec(userSchema)
	Db.MustExec(eventSchema)
}
