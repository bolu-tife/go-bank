package main

import (
	"flag"
	"fmt"
	"log"
)

func main() {
	seed := flag.Bool("seed", false, "seed the db")
	flag.Parse()

	store, err := NewPostgressStore()

	if err != nil {
		log.Fatal(err)
	}

	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	if *seed {
		fmt.Println("seeding the database")
		seedAccounts(store)
	}

	port := fmt.Sprintf(":%s", goDotEnvVariable("PORT"))
	server := NewApiServer(port, store)
	server.Run()
}
