package main

import (
	"context"
	"fmt"
	"os"

	"github.com/PPRAMANIK62/devhunt/internal/config"
	"github.com/PPRAMANIK62/devhunt/internal/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	db, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "database error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	fmt.Printf("connected to postgres\nport=%s env=%s\n", cfg.ServerPort, cfg.Env)
}
