package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type APIServer struct {
	listenAddr string
	store      Storage
}

func NewApiServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

func goDotEnvVariable(key string) string {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file")
	}
	return os.Getenv(key)
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/login", makeHTTPHandleFunc(s.handleLogin))
	router.HandleFunc("/account", makeHTTPHandleFunc(s.handleAccount))
	router.HandleFunc("/account/{id}", withJWTAuth(makeHTTPHandleFunc(s.handleAccountById), s.store))
	router.HandleFunc("/transfer", makeHTTPHandleFunc(s.handleTransfer))
	router.HandleFunc("/account/{id}/reactivate", makeHTTPHandleFunc(s.handleAccountReactivation))

	log.Println("JSON API server running on port: ", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method not allowed %s", r.Method)
	}
	req := new(LoginRequest)

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return err
	}
	account, err := s.store.GetAccountByEmail(req.Email)
	if err != nil || !account.ValidatePassword(req.Password) {
		return fmt.Errorf("email or password is wrong")
	}

	tokenString, err := createJWT(account)
	if err != nil {
		return err
	}

	res := LoginResponse{
		Token: tokenString,
		Email: account.Email,
	}

	return WriteJSON(w, http.StatusOK, res)
}

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return s.handleGetAccount(w, r)
	case "POST":
		return s.handleCreateAccount(w, r)
	default:
		return fmt.Errorf("method not allowed %s", r.Method)
	}
}

func (s *APIServer) handleAccountById(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return s.handleGetAccountByID(w, r)
	case "DELETE":
		return s.handleDeleteAccount(w, r)
	case "PATCH":
		return s.handleUpdateAccountDetails(w, r)
	default:
		return fmt.Errorf("method not allowed %s", r.Method)
	}
}

func (s *APIServer) handleAccountReactivation(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "PATCH":
		return s.handleReactivateAccount(w, r)
	default:
		return fmt.Errorf("method not allowed %s", r.Method)
	}
}

func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "POST":
		return s.handleMoneyTransfer(w, r)
	default:
		return fmt.Errorf("method not allowed %s", r.Method)
	}
}

func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
	skip, _ := strconv.Atoi(r.URL.Query().Get("skip"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if skip < 0 {
		skip = 0
	}

	if limit < 1 || limit > 10 {
		limit = 10
	}

	accounts, err := s.store.GetAccounts(skip, limit)
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, accounts)
}

func (s *APIServer) handleGetAccountByID(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err
	}

	account, err := s.store.GetAccountByID(id)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	createAccountReq := new(AccountRequest)

	if err := json.NewDecoder(r.Body).Decode(createAccountReq); err != nil {
		return err
	}
	account, err := NewAccount(createAccountReq.FirstName, createAccountReq.LastName, createAccountReq.Email, createAccountReq.Password)
	if err != nil {
		return err
	}
	if err := s.store.CreateAccount(account); err != nil {
		return err
	}

	tokenString, err := createJWT(account)
	if err != nil {
		return err
	}

	fmt.Println("token sting", tokenString)
	return WriteJSON(w, http.StatusCreated, account)
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err
	}
	if err := s.store.DeleteAccount(id); err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, "Deleted successfully")
}

func (s *APIServer) handleReactivateAccount(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err
	}
	if err := s.store.ReactivateAccount(id); err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, "Account reactivated successfully")
}

func (s *APIServer) handleUpdateAccountDetails(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err
	}

	updateAccountReq := new(AccountRequest)
	if err := json.NewDecoder(r.Body).Decode(updateAccountReq); err != nil {
		return err
	}

	if err := s.store.UpdateAccountDetails(id, updateAccountReq.FirstName, updateAccountReq.LastName); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, "Account updated successfully")
}

func (s *APIServer) handleMoneyTransfer(w http.ResponseWriter, r *http.Request) error {
	transferReq := new(TransferRequest)
	if err := json.NewDecoder(r.Body).Decode(transferReq); err != nil {
		return err
	}

	defer r.Body.Close()
	return WriteJSON(w, http.StatusOK, transferReq)
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func createJWT(account *Account) (string, error) {

	// Create claims while leaving out some of the optional fields
	claims := &jwt.MapClaims{
		"expiresAt":     15000,
		"accountNumber": account.Number,
		"issuer":        "test",
	}

	jwtSecret := goDotEnvVariable("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(jwtSecret))
}

func permissionDenied(w http.ResponseWriter) {
	WriteJSON(w, http.StatusForbidden, APIError{Error: "You're Unauthorized!"})
}

func withJWTAuth(handlerFunc http.HandlerFunc, s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Calling JWT auth middleware")

		tokenString := r.Header.Get("Authorization")

		token, err := validateJWT(tokenString)
		if !token.Valid {
			permissionDenied(w)
			return
		}

		if err != nil {
			permissionDenied(w)
			return
		}

		userID, err := getID(r)
		if err != nil {
			permissionDenied(w)
			return
		}

		account, err := s.GetAccountByID(userID)
		if err != nil {
			permissionDenied(w)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		if account.Number != int64(claims["accountNumber"].(float64)) {
			permissionDenied(w)
			return
		}

		handlerFunc(w, r)
	}
}

func validateJWT(tokenString string) (*jwt.Token, error) {
	jwtSecret := goDotEnvVariable("JWT_SECRET")

	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(jwtSecret), nil
	})
}

type APIError struct {
	Error string `json:"error"`
}

type apiFunc func(http.ResponseWriter, *http.Request) error

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			// handle the error
			WriteJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		}
	}
}

func getID(r *http.Request) (int, error) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return id, fmt.Errorf("invalid id given %s", idStr)
	}
	return id, nil
}
