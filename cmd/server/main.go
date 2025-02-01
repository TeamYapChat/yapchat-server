package main

import (
	"fmt"
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
		log.Error.Fatalln("Error while connecting to database:", err)
	}
}

func main() {
	fmt.Println("Howdy!")
}
