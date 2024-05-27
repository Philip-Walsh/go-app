package main

import (
	"bytes"
	// "fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/Philip-Walsh/go-app/news"
	"github.com/joho/godotenv"
)

var tpl = template.Must(template.ParseFiles("index.html"))

func indexHandler(w http.ResponseWriter, r *http.Request) {
	buf := &bytes.Buffer{}
	err := tpl.Execute(buf, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	buf.WriteTo(w)
}

func (s *Search) IsLastPage() bool {
	return s.NextPage >= s.TotalPages
}
func (s *Search) CurrentPage() int {
	if s.NextPage == 1 {
		return s.NextPage
	}

	return s.NextPage - 1
}
func (s *Search) PreviousPage() int {
	return s.CurrentPage() - 1
}

func searchHandler(newsapi *news.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := url.Parse(r.URL.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		params := u.Query()
		searchQuery := params.Get("q")
		page := params.Get("page")
		if page == "" {
			page = "1"
		}

		results, err := newsapi.FetchEverything(searchQuery, page)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		nextPage, err := strconv.Atoi(page)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		search := &Search{
			Query:      searchQuery,
			NextPage:   nextPage,
			TotalPages: int(math.Ceil(float64(results.TotalResults) / float64(newsapi.PageSize))),
			Results:    results,
		}
		if ok := !search.IsLastPage(); ok {
			search.NextPage++
		}
		buf := &bytes.Buffer{}
		err = tpl.Execute(buf, search)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		buf.WriteTo(w)
	}
}

var newsapi *news.Client

type Search struct {
	Query      string
	NextPage   int
	TotalPages int
	Results    *news.Results
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
	apiKey := os.Getenv("NEWS_API_KEY")
	if apiKey == "" {
		log.Fatal("Env: NEWS_API_KEY must be set")
	}

	myClient := &http.Client{Timeout: 10 * time.Second}
	newsapi := news.NewClient(myClient, apiKey, 20)

	fileServer := http.FileServer(http.Dir("assets"))
	// Router
	mux := http.NewServeMux()

	mux.HandleFunc("/search", searchHandler(newsapi))
	mux.HandleFunc("/", indexHandler)
	mux.Handle("/assets/", http.StripPrefix("/assets/", fileServer))
	http.ListenAndServe(":"+port, mux)
	log.Println("Listening on port ", port)
}
