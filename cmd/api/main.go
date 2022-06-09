package main

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"strings"
	"sync"
	"time"

	elasticsearch "github.com/elastic/go-elasticsearch/v7"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hafizmfadli/hello-nerds-api/internal/data"
	"github.com/hafizmfadli/hello-nerds-api/internal/jsonlog"
	"github.com/hafizmfadli/hello-nerds-api/internal/mailer"
)

const version = "1.0.0"

type config struct {
	port int
	env string
	db struct {
		dsn string
	}
	es elasticsearch.Config
	smtp struct {
		host string
		port int
		username string
		password string
		sender string
	}
}

type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg sync.WaitGroup
	temp struct {
		checkoutCounter int
		adminUpdateCounter int
	}
}

func main() {

	var cfg config
	var clusterURLs string
	
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("HELLO_NERDS_DB_DSN"), "MySQL DSN")
	flag.StringVar(&clusterURLs, "es-cluster-URLs", "http://127.0.0.1:9200", "Elasticsearch Cluster URLs")

	// Read the SMTP server configuration settings into the config struct, using the
	// Mailtrap settings as the default values.
	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 587, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "4252e4b90aa4cd", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "2924c5b1acf3a9", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Hello Nerds <no-reply@hello.nerds.net>", "SMTP sender")

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
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	// Call app.serve() to start the server
	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}

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

	_, err = db.ExecContext(ctx, "SET GLOBAL TRANSACTION ISOLATION LEVEL REPEATABLE READ")
	if err != nil {
		return nil, err
	}

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