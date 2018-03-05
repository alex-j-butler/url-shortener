package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	muxtrace "github.com/DataDog/dd-trace-go/contrib/gorilla/mux"
	"github.com/gorilla/mux"
	hashids "github.com/speps/go-hashids"
)

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("postgres", "user=pqgotest dbname=pqgotest sslmode=verify-full")
	if err != nil {
		log.Fatal(err)
	}

	r := muxtrace.NewRouter(muxtrace.WithServiceName("url-shortener.mux"))
	r.HandleFunc("/create", ShortenHandler).Queries("url", "")
	r.HandleFunc("/{hashid:[a-zA-Z0-9]+}", ShortenedHandler)
	r.HandleFunc("/", CatchAllHandler)

	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}

func ShortenHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	var id int
	err := db.QueryRow("INSERT INTO urls (url) VALUES ($1) RETURNING id", url).Scan(&id)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	hd := hashids.NewData()
	hd.Salt = "url-shortener"
	h, _ := hashids.NewWithData(hd)
	hashid, _ := h.Encode([]int{id})

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("http://qix.tf/%s", hashid)))
}

func ShortenedHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hashid, ok := vars["hashid"]
	if !ok {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	hd := hashids.NewData()
	hd.Salt = "url-shortener"
	h, _ := hashids.NewWithData(hd)
	ids := h.Decode(hashid)
	if len(ids) == 0 {
		http.NotFound(w, r)
		return
	}
	id := ids[0]

	var url string
	err := db.QueryRow("SELECT url FROM urls WHERE id = $1", id).Scan(&url)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, url, http.StatusMovedPermanently)
}

func CatchAllHandler(w http.ResponseWriter, r *http.Request) {

}
