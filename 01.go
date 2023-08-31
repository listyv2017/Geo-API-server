package main

import (
	"fmt"
	"net/http"
)

func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Hello\n")
}

func main() {
	http.HandleFunc("/", hello)

	serverAddr := ":8080"
	fmt.Printf("Server listening on %s...\n", serverAddr)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Error starting server: %v", err)
	}
}
