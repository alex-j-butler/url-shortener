package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	hashids "github.com/speps/go-hashids"
	"github.com/spf13/viper"
)

type ShortenOne struct {
	URL string `json:"url"`
}

type ShortenMultiple struct {
	URLs []string `json:"urls"`
}

func ShortenHandler(w http.ResponseWriter, r *http.Request) error {
	var req ShortenOne
	j := json.NewDecoder(r.Body)
	err := j.Decode(&req)
	if err != nil {
		return err
	}

	var id int
	err = db.QueryRow("INSERT INTO urls (url) VALUES ($1) RETURNING id", req.URL).Scan(&id)
	if err != nil {
		return errors.New("Database error")
	}

	hd := hashids.NewData()
	hd.Salt = "url-shortener"
	h, _ := hashids.NewWithData(hd)
	hashid, _ := h.Encode([]int{id})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	resp := ShortenOne{
		URL: fmt.Sprintf("%s/%s", viper.GetString("base_url"), hashid),
	}

	// Write JSON response.
	enc := json.NewEncoder(w)
	err = enc.Encode(resp)
	if err != nil {
		return err
	}
	return nil
}

func ShortenMultipleHandler(w http.ResponseWriter, r *http.Request) error {
	var req ShortenMultiple
	j := json.NewDecoder(r.Body)
	err := j.Decode(&req)
	if err != nil {
		return err
	}

	ids := make([]int, 0, len(req.URLs))
	urls := make([]string, 0, len(req.URLs))

	db.Exec("BEGIN")
	for _, url := range req.URLs {
		var id int
		err := db.QueryRow("INSERT INTO urls (url) VALUES ($1) RETURNING id", url).Scan(&id)
		if err != nil {
			return errors.New("Database error")
		}

		ids = append(ids, id)
	}
	db.Exec("COMMIT")

	for _, id := range ids {
		hd := hashids.NewData()
		hd.Salt = "url-shortener"
		h, _ := hashids.NewWithData(hd)
		hashid, _ := h.Encode([]int{id})

		urls = append(urls, fmt.Sprintf("%s/%s", viper.GetString("base_url"), hashid))
	}

	var resp ShortenMultiple
	resp.URLs = urls

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Write JSON response.
	enc := json.NewEncoder(w)
	err = enc.Encode(resp)
	if err != nil {
		return err
	}

	return nil
}

func ShortenedHandler(w http.ResponseWriter, r *http.Request) {
	cook, err := r.Cookie("qix")
	if err == http.ErrNoCookie {
		str, err := GenerateRandomString(64)
		if err == nil {
			expiration := time.Now().Add(365 * 24 * time.Hour)
			cookie := &http.Cookie{Name: "qix", Value: str, Expires: expiration}
			http.SetCookie(w, cookie)
			cook = cookie
		}
	}

	vars := mux.Vars(r)
	hashid, _ := vars["hashid"]

	hd := hashids.NewData()
	hd.Salt = "url-shortener"
	h, _ := hashids.NewWithData(hd)
	ids, err := h.DecodeWithError(hashid)
	if err != nil {
		http.Redirect(w, r, viper.GetString("default_url"), http.StatusMovedPermanently) // Redirect to default URL
		return
	}
	if len(ids) == 0 {
		http.Redirect(w, r, viper.GetString("default_url"), http.StatusMovedPermanently) // Redirect to default URL
		return
	}
	id := ids[0]

	var url string
	err = db.QueryRow("SELECT url FROM urls WHERE id = $1", id).Scan(&url)
	if err != nil {
		http.Redirect(w, r, viper.GetString("default_url"), http.StatusMovedPermanently) // Redirect to default URL
		return
	}

	if cook != nil {
		redisClient.PFAdd(fmt.Sprintf("url:%d", id), cook.Value)
	}
	http.Redirect(w, r, url, http.StatusMovedPermanently)
}
