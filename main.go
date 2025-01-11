package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

func getDBUrl() string {
	envFile, err := os.ReadFile(".env")
	if err != nil {
		log.Fatal(err)
	}
	env := string(envFile)
	url := strings.Split(env, "=")[1]
	url = url[1 : len(url)-1]
	return url
}

var schema = `
CREATE TABLE Files(
	id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
	name text,
	content text,
	author text
	);
`

func pushSchema(db *sqlx.DB) {
	db.MustExec(schema)
}

func dbConnect() *sqlx.DB {
	db, err := sqlx.Connect("postgres", getDBUrl())
	if err != nil {
		log.Fatalln(err)
	}
	return db
}

type File struct {
	ID      string `db:"id" json:"id"`
	Name    string `db:"name" json:"name"`
	Content string `db:"content" json:"content"`
	Author  string `db:"author" json:"author"`
}

func main() {
	db := dbConnect()
	defer db.Close()
	e := echo.New()
	e.GET("/file/:id", func(c echo.Context) error {
		file := new(File)
		if err := db.Get(file, "SELECT * FROM Files WHERE id = $1", c.Param("id")); err != nil {
			c.Echo().Logger.Print(err)
			if sql.ErrNoRows == err {
				return echo.NewHTTPError(http.StatusNotFound, "File Not found")
			} else {
				return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
			}
		}
		return c.JSON(http.StatusOK, file)
	})
	e.POST("/file", func(c echo.Context) error {
		file := new(File)
		if err := c.Bind(file); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Bad Request")
		}
		if _, err := db.Exec("INSERT INTO Files (name, content, author) VALUES ($1, $2, $3)", file.Name, file.Content, file.Author); err != nil {
			c.Echo().Logger.Print(err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
		}
		return c.JSON(http.StatusCreated, echo.Map{"msg": "File Created"})
	})
	e.Logger.Fatal(e.Start(":1323"))
}
