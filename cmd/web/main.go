package main

import (
	"log"
	"net/http"
)

const webPort = "8080"

func main() {
	mux := routes()

	log.Println("Starting web server on port", webPort)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
