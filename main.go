package main

import (
	"fmt"
	"net/http"
	"log"
)

func main () {
	fmt.Println("Hello World")
	servM := http.NewServeMux()
	servM.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	servM.HandleFunc("/healthz", ContentTypeHandler)
	server := http.Server {
		Handler: servM,
		Addr: ":8080",
	}
	
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

}

func ContentTypeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}
