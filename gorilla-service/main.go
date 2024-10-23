package main

import (
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func keyValueGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	value, err := Get(key)
	if errors.Is(err, ErrorNoSuchKey) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(value))

	log.Printf("GET key=%s\n", key)
}

func keyValuePutHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	value, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	err = Put(key, string(value))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

	log.Printf("PUT key=%s value=%s\n", key, string(value))
}

func main() {
	// Create a new mux router
	r := mux.NewRouter()

	// Register putHandler as the handler function for PUT
	// requests matching "/v1/key/{key}"
	r.HandleFunc("/v1/key/{key}", keyValuePutHandler).Methods("PUT")

	// Register putHandler as the handler function for GET
	// requests matching "/v1/key/{key}"
	r.HandleFunc("/v1/key/{key}", keyValueGetHandler).Methods("GET")

	// Bind to a port and pass in the mux router
	log.Fatal(http.ListenAndServe(":8080", r))
}
