package main

import (
	"bms/server/app"
	_ "github.com/lib/pq"
	_ "strconv"
)

func main() {
	config := app.Config{
		Host:       "localhost",
		DbPort:     "5432",
		DbUser:     "postgres",
		DbPassword: "password",
		DbName:     "bms_db",
		ServerPort: "8080",
	}

	app := app.NewApp(config)
	app.Run()
}
