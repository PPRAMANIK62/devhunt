package main

import (
	"fmt"
	"os"

	"github.com/PPRAMANIK62/devhunt/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("config loaded --port=%s env=%s\n", cfg.ServerPort, cfg.Env)
}
