package main

import (
	"net/http"
)

func main() {
	servmux := http.NewServeMux()

	server := &http.Server{Handler: servmux, Addr: ":8080"}

	err := server.ListenAndServe()
	if err != nil {
		// It's good to handle errors
		panic(err)
	}

}
