package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	elasticsearch "github.com/elastic/go-elasticsearch/v7"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hafizmfadli/hello-nerds-api/internal/data"
)

const version = "1.0.0"

type config struct {
	port int
	env string
	db struct {
		dsn string
	}
	es elasticsearch.Config
}

type application struct {
	config config
	logger *log.Logger
	models data.Models
}

func main() {

	var cfg config
	var clusterURLs string
	
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("HELLO_NERDS_DB_DSN"), "MySQL DSN")
	flag.StringVar(&clusterURLs, "es-cluster-URLs", "http://127.0.0.1:9200", "Elasticsearch Cluster URLs")
	flag.Parse()
	cfg.es.Addresses = strings.Split(clusterURLs, ",")
	

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// create connection pool
	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	logger.Printf("database connection pool established")

	// create es connection
	es, err := openES(cfg)
	if err != nil {
		logger.Fatal(err)
	}

	// inject all dependencies
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModel(db, es),
	}

	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.port),
		Handler: app.routes(),
		IdleTimeout: time.Minute,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Printf("starting %s server on %s", cfg.env, srv.Addr)
	err = srv.ListenAndServe()
	log.Fatal(err)
}

// The openDB() function returns a sql.DB connection pool
func openDB(cfg config) (*sql.DB, error) {

	db, err := sql.Open("mysql", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// create  a context with a 5-second timeout deadline
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()


	// use PingContext() to establish a new connection to database, passing in the
	// context we created aboe as parameter. If the connection couldn't be
	// established successfully within 5 second deadline, then this will return an
	// error
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// The openES() function returns a elasticsearch.Client connection
func openES(cfg config) (*elasticsearch.Client, error) {
	es, err := elasticsearch.NewClient(cfg.es)
	if err != nil {
		return nil, err
	}
	return es, nil
}