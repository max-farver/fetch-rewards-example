package main

import (
	"context"
	"database/sql"
	"fetch-rewards/server"
	"fetch-rewards/services"
	"fmt"
	"github.com/go-playground/validator/v10"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
	"log"
)

func main() {
	fmt.Println("30")

	ctx, contextCancel := context.WithCancel(context.Background())
	defer contextCancel()

	db, err := sql.Open("sqlite3", "fetch.db")
	if err != nil {
		log.Fatal(err.Error())
	}

	dbSchema := `
		CREATE TABLE IF NOT EXISTS fetch_rewards (
			payer varchar(255) NOT NULL,
			points int NOT NULL,
			timestamp datetime NOT NULL
		)`

	_, err = db.Exec(dbSchema)
	if err != nil {
		log.Printf("%q: %s\n", err, dbSchema)
		return
	}

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	sugar := logger.Sugar()

	api := &server.Server{
		Ctx: ctx,
		DB: db,
		Logger: sugar,
		Validator: validator.New(),

		PointsService: &services.PointsService{Ctx: ctx, DB: db},
	}
	api.InitServer()
}