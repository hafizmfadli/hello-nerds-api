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
	"github.com/hafizmfadli/hello-nerds-api/internal/jsonlog"
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
	logger *jsonlog.Logger
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
	
	// Initialize a new jsonlog.Logger which writes any messages *at or above* the INFO
	// severity level to the standard out stream
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	// create connection pool
	db, err := openDB(cfg)
	if err != nil {
		// Use the PrintFatal() method to write a log entry containing the error at the
		// FATAL level and exit. We have no additional properties to include in the log
		// entry, so we pass nil as the second parameter
		logger.PrintFatal(err, nil)
	}
	defer db.Close()

	// Like wise use PrintInfo() method to write a message at the INFO level
	logger.PrintInfo("database connection pool established", nil)

	// create es connection
	es, err := openES(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
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
		// Create a new GO log.Logger instance with the log.New() function, passing in
		// our custom Logger as the first parameter. The "" and 0 indicate that the
		// log.Logger instance should not use a prefix or any flags.
		ErrorLog: log.New(logger, "", 0),
		IdleTimeout: time.Minute,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env": cfg.env,
	})

	err = srv.ListenAndServe()

	logger.PrintFatal(err, nil)
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