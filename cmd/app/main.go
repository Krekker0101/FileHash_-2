package main

import (
	"fmt"
	"os"

	"github.com/filehash/internal/app"
	"github.com/filehash/internal/config"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	_ = godotenv.Load(".env.example")

	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("config load error: %w", err))
	}

	application, err := app.New(cfg)
	if err != nil {
		panic(fmt.Errorf("app initialization error: %w", err))
	}
	defer func() {
		if err := application.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "error closing app: %v\n", err)
		}
	}()

	if err := application.Run(); err != nil {
		panic(fmt.Errorf("app run error: %w", err))
	}
}
