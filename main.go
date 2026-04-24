//go:build js && wasm

package main

import (
	"database/sql"
	"log"
	"net/http"

	flagapi "github.com/sophic00/cf-flag.git/internal"
	"github.com/syumai/workers"
	"github.com/syumai/workers/cloudflare"
	_ "github.com/syumai/workers/cloudflare/d1"
)

func main() {
	hashSecret := cloudflare.Getenv("FLAG_HASH_KEY")
	if hashSecret == "" {
		hashSecret = "dev-only-secret"
	}

	db, err := sql.Open("d1", "DB")
	if err != nil {
		log.Fatalf("open d1: %v", err)
	}

	server := flagapi.New(db, hashSecret)
	mux := http.NewServeMux()
	server.RegisterRoutes(mux)
	workers.Serve(mux)
}
