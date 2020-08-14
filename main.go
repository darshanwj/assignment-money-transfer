package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/darshanwj/assignment-money-transfer/service"

	"github.com/gorilla/mux"
)

func main() {
	h := &handler{service.New()}
	r := mux.NewRouter()
	r.Use(commonMiddleware)
	r.HandleFunc("/transfer", h.TransferHandler).Methods(http.MethodPost)
	r.HandleFunc("/accounts", h.GetAccountsHandler).Methods(http.MethodGet)
	r.HandleFunc("/accounts", h.CreateAccountHandler).Methods(http.MethodPost)
	log.Fatal(http.ListenAndServe(":8090", r))
}

type handler struct {
	service.Service
}

// Reads all accounts and outputs JSON
func (h *handler) GetAccountsHandler(w http.ResponseWriter, r *http.Request) {
	_ = json.NewEncoder(w).Encode(h.GetAccounts())
}

// Creates a single account and outputs JSON
func (h *handler) CreateAccountHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var acc service.Account
	if err = json.Unmarshal(body, &acc); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	h.CreateAccount(acc)

	_ = json.NewEncoder(w).Encode(acc)
}

// Performs a transfer from one account to another
func (h *handler) TransferHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var trn service.Transfer
	if err = json.Unmarshal(body, &trn); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	acc, err := h.Transfer(trn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_ = json.NewEncoder(w).Encode(acc)
}

// Sets JSON content type header for all responses
func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
