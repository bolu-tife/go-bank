package main

import "log"

func seedAccount(store Storage, firstName, lastName, email, password string) *Account {
	acc, err := NewAccount(firstName, lastName, email, password)
	if err != nil {
		log.Fatal(err)
	}

	if err := store.CreateAccount(acc); err != nil {
		log.Fatal(err)
	}

	return acc
}

func seedAccounts(s Storage) {
	seedAccount(s, "Test", "Tester", goDotEnvVariable("SEED_EMAIL"), goDotEnvVariable("SEED_PASSWORD"))
}
