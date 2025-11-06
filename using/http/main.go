package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	// @idiomatic: using Fprintf to write string to io.Writer
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintf(w, time.Now().Format(time.RFC3339))
		if err != nil {
			log.Printf("Error: %v", err)
		}
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}
