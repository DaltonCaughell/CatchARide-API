package middleware

import (
	"database/sql"
	"github.com/go-martini/martini"
	_ "github.com/go-sql-driver/mysql"
)

func BasicAuth(db *sql.DB, c martini.Context) {

}
