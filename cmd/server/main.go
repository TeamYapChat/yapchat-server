package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/teamyapchat/yapchat-server/internal/database"
	log "github.com/teamyapchat/yapchat-server/internal/logging"
)

func init() {
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	host := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user,
		pass,
		host,
		dbName,
	)

	_, err := database.Connect(dsn)
	if err != nil {
		panic(err)
	}
	log.Info.Println("Successfully connected to database")
}

func main() {
	fs := http.FileServer(http.Dir("./dist"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := os.Stat("./dist" + r.URL.Path)
		if os.IsNotExist(err) {
			r.URL.Path = "/"
		}

		fs.ServeHTTP(w, r)
	})

	http.ListenAndServe(":8081", nil)
}
