package main

import (
	"fmt"
	"net/http"
	"log"
)

func main () {
	fmt.Println("Hello World")
	servM := http.NewServeMux()
	servM.Handle("/assets/logo.png", http.FileServer(http.Dir(".")))
	server := http.Server {
		Handler: servM,
		Addr: ":8080",
	}
	
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

}
