package main

import (
	"log"
	"net/http"
)

func main() {
	const port = "8081"

	mux := http.NewServeMux()

	srv := &http.Server{
		Addr:    "127.0.0.1:" + port,
		Handler: mux,
	}

	mux.Handle("/", http.FileServer(http.Dir("./")))

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())

}
