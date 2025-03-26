package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("Starting WEX TAG Transaction Processing System")

	// TODO: Initialize application components

	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
