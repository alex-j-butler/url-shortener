package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"alex-j-butler.com/url-shortener/config"
	sqltrace "github.com/DataDog/dd-trace-go/contrib/database/sql"
	redistrace "github.com/DataDog/dd-trace-go/contrib/go-redis/redis"
	muxtrace "github.com/DataDog/dd-trace-go/contrib/gorilla/mux"
	"github.com/fsnotify/fsnotify"
	"github.com/go-redis/redis"
	"github.com/spf13/viper"

	pq "github.com/lib/pq"
)

var db *sql.DB
var redisClient *redistrace.Client

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalln(err)
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Println("Reloaded configuration file")
	})

	viper.SetDefault("postgres_dsn", "dbname=urlshortener sslmode=disable host=/run/postgresql")
	viper.SetDefault("redis_address", "localhost:6379")
	viper.SetDefault("redis_password", "")
	viper.SetDefault("redis_db", 0)
	viper.SetDefault("bind_address", "127.0.0.1")
	viper.SetDefault("bind_port", 8080)

	sqltrace.Register("postgres", &pq.Driver{}, sqltrace.WithServiceName("url-shortener.db"))
	db, err = sqltrace.Open("postgres", viper.GetString("postgres_dsn"))
	if err != nil {
		log.Fatal(err)
	}

	redisClient = redistrace.NewClient(&redis.Options{
		Addr:     viper.GetString("redis_address"),
		Password: viper.GetString("redis_password"),
		DB:       viper.GetInt("redis_db"),
	})

	r := muxtrace.NewRouter(muxtrace.WithServiceName("url-shortener.mux"))
	r.Handle("/create", APIHandler{Handler: ShortenHandler}).Methods("POST")
	r.Handle("/createMultiple", APIHandler{Handler: ShortenMultipleHandler}).Methods("POST")
	r.HandleFunc("/{hashid:[a-zA-Z0-9]+}", ShortenedHandler)
	r.HandleFunc("/", CatchAllHandler)

	http.Handle("/", r)
	http.ListenAndServe(fmt.Sprintf("%s:%d", viper.GetString("bind_address"), viper.GetInt("bind_port")), nil)
}

func CatchAllHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, config.Conf.DefaultURL, http.StatusMovedPermanently)
}
