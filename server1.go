package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	router := http.NewServeMux()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"message": "Hello from server 1"})
	})

	server := http.Server{
		Addr:    ":4001",
		Handler: router,
	}

	log.Printf("serving requests at 'localhost:4001'")
	log.Fatal(server.ListenAndServe())
}
