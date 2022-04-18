# Hello Nerds API

## How to Run

### 1.) Setup environment variable

Set MySQL dsn (data source name) as environment variable with name HELLO_NERDS_DB_DSN

Example for Linux distribution or MacOS  :

`export HELLO_NERDS_DB_DSN="root:debezium@tcp(localhost:3307)/inventory?parseTime=true"`

### 2.) Install all dependencies

In root project directory run this command : 

`go mod download`

### 3.) Run

Run this command to see available command's flag in this project :

`go run ./cmd/api/* -h`

You will see the output like this :

```
Usage of /tmp/go-build994341346/b001/exe/books:

  -db-dsn string
        MySQL DSN (default "root:debezium@tcp(localhost:3307)/inventory?parseTime=true")


  -env string
        Environment (development|staging|production) (default "development")


  -es-cluster-URLs string
        Elasticsearch Cluster URLs (default "http://127.0.0.1:9200")


  -port int
        API server port (default 4000)
```

Note : make sure you pass appropriate value for those flags or the project will be failed to run.