package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	store, err := NewPostgressStore()

	if err != nil {
		log.Fatal(err)
	}

	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	port := fmt.Sprintf(":%s", os.Getenv("PORT"))
	server := NewApiServer(port, store)
	server.Run()
}
