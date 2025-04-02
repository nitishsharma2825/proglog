package main

import (
	"log"

	"github.com/nitishsharma2825/proglog/internal/server"
)

func main() {
	srv := server.NewHTTPServer(":8080")
	log.Fatal(srv.ListenAndServe())
}
