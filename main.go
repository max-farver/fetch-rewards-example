package main

import (
	"context"
	"database/sql"
	"fetch-rewards/server"
	"fetch-rewards/services"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"log"
	_ "modernc.org/sqlite"
	"os"
	"path"
)

func main() {
	ctx, contextCancel := context.WithCancel(context.Background())
	defer contextCancel()

	db, err := sql.Open("sqlite", "fetch.db")
	if err != nil {
		log.Fatal(err.Error())
	}

	dbSchema := `
		CREATE TABLE IF NOT EXISTS fetch_rewards (
		    id INTEGER PRIMARY KEY AUTOINCREMENT,
			payer VARCHAR(255) NOT NULL,
			remaining_points INTEGER NOT NULL,
			timestamp DATETIME NOT NULL,
			points INTEGER NOT NULL
		)`

	_, err = db.Exec(dbSchema)
	if err != nil {
		log.Printf("%q: %s\n", err, dbSchema)
		return
	}

	// Setup logger
	cwd, err := os.Getwd()
	logDirPath := path.Join(cwd, "logs")
	err = os.Mkdir(logDirPath, 0755)
	if err != nil {
		log.Printf(err.Error())
	}

	logFilePath := path.Join(logDirPath, "example.log")
	_, err = os.Create(logFilePath)
	if err != nil {
		log.Printf(err.Error())
	}

	cfg := zap.NewDevelopmentConfig()
	cfg.OutputPaths = []string{logFilePath}
	logger, err := cfg.Build()
	if err != nil {
		log.Printf("%q: %s\n", err, dbSchema)
		return
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	api := &server.Server{
		Ctx:       ctx,
		DB:        db,
		Logger:    sugar,
		Validator: validator.New(),

		PointsService: &services.PointsService{Ctx: ctx, DB: db},
	}
	api.InitServer()
}
