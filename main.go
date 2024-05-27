package main

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var tpl = template.Must(template.ParseFiles("index.html"))

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tpl.Execute(w, nil)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	fileServer := http.FileServer(http.Dir("assets"))
	// Router
	mux := http.NewServeMux()

	mux.HandleFunc("/", indexHandler)
	mux.Handle("/assets/", http.StripPrefix("/assets/", fileServer))
	http.ListenAndServe(":"+port, mux)
	log.Println("Listening on port ", port)
}
