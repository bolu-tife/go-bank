package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	ReactivateAccount(int) error
	UpdateAccount(*Account) error
	GetAccountByID(int) (*Account, error)
	GetAccounts(int, int) ([]*Account, error)
	UpdateAccountDetails(int, string, string) error
}

type PostgressStore struct {
	db *sql.DB
}

func NewPostgressStore() (*PostgressStore, error) {
	username := goDotEnvVariable("APP_DB_USERNAME")
	password := goDotEnvVariable("APP_DB_PASSWORD")
	db_name := goDotEnvVariable("APP_DB_NAME")
	ssl_mode := goDotEnvVariable("SSL_MODE")

	connStr := fmt.Sprintf("user=%s dbname=%s password=%s sslmode=%s", username, db_name, password, ssl_mode)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &PostgressStore{
		db: db,
	}, nil
}

func (s *PostgressStore) Init() error {
	return s.CreateAccountTable()
}

func (s *PostgressStore) CreateAccountTable() error {
	query := `create table if not exists account(
		id serial primary key,
		first_name varchar(50),
		last_name varchar(50),
		number serial,
		balance serial,
		created_at timestamp,
		deleted boolean not null
	)`

	_, err := s.db.Exec(query)
	return err
}

func (s *PostgressStore) CreateAccount(acc *Account) error {
	query := `insert into account
	(first_name, last_name, number, balance, created_at)
	values ($1, $2, $3, $4, $5)`
	resp, err := s.db.Query(
		query,
		acc.FirstName,
		acc.LastName,
		acc.Number,
		acc.Balance,
		acc.CreatedAt,
	)

	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", resp)

	return nil
}

func (s *PostgressStore) DeleteAccount(id int) error {
	_, err := s.db.Query("update account set deleted=true where id=$1 and deleted=false", id)
	return err
}

func (s *PostgressStore) ReactivateAccount(id int) error {
	_, err := s.db.Query("update account set deleted=false where id=$1 and deleted=true", id)
	return err
}

func (s *PostgressStore) UpdateAccountDetails(id int, firstName, lastName string) error {
	_, err := s.db.Query("update account set first_name=$1, last_name=$2 where id=$3 and deleted=false", firstName, lastName, id)
	return err
}

func (s *PostgressStore) UpdateAccount(*Account) error {
	return nil
}

func (s *PostgressStore) GetAccountByID(id int) (*Account, error) {
	rows, err := s.db.Query("select * from account where id = $1 ", id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		account, err := scanIntoAccount(rows)

		if err != nil {
			return nil, err
		}

		if account.Deleted {
			return nil, fmt.Errorf("account %d is deactivated", id)
		}

		return account, nil
	}

	return nil, fmt.Errorf("account %d not found", id)
}

func (s *PostgressStore) GetAccounts(skip, limit int) ([]*Account, error) {
	query := "select * from account where deleted=false limit $1 offset $2"
	rows, err := s.db.Query(query, limit, skip)

	if err != nil {
		return nil, err
	}
	accounts := []*Account{}

	for rows.Next() {
		account, err := scanIntoAccount(rows)

		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}
	return accounts, nil
}

func scanIntoAccount(rows *sql.Rows) (*Account, error) {
	account := new(Account)
	err := rows.Scan(
		&account.ID,
		&account.FirstName,
		&account.LastName,
		&account.Number,
		&account.Balance,
		&account.CreatedAt,
		&account.Deleted,
	)

	return account, err
}
