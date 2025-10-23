package main

import (
	"log"
	"net/http"

	api "github.com/kebab0o/sysdash/backend/internal/http"
	"github.com/kebab0o/sysdash/backend/internal/store"
)

func main() {
	mem := store.NewMemory()
	app := &api.App{Store: mem}
	srv := api.NewServer(app.Routes())

	log.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", srv.Router); err != nil {
		log.Fatal(err)
	}
}
