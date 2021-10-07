package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	httpHandler "github.com/bxcodec/aqua/handler"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq" //import for postgres driver

	_ "net/http/pprof" // import for profiling
)

func main() {
	dbHost := viper.GetString("postgres.host")
	dbPort := viper.GetString("postgres.port")
	dbUser := viper.GetString("postgres.user")
	dbPass := viper.GetString("postgres.pass")
	dbName := viper.GetString("postgres.name")
	// dsn := dbUser + `:` + dbPass + `@tcp(` + dbHost + `:` + dbPort + `)/` + dbName + `?sslmode=disable`
	dsn := "postgresql://" + dbUser + ":" + dbPass + "@" + dbHost + ":" + dbPort + "/" + dbName + "?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Minute * 10)

	handler := httpHandler.InitArticle(db)
	echoServer := echo.New()

	// Register the handler
	echoServer.GET("/articles", handler.FetchArticles)

	errCh := make(chan error)

	go func(ch chan error) {
		log.Println("Starting HTTP server")
		errCh <- echoServer.Start(":9090")
	}(errCh)

	go func(ch chan error) {
		log.Println("Starting Profiling HTTP server")
		errCh <- http.ListenAndServe(":8080", nil)
	}(errCh)

	for {
		log.Fatal(<-errCh)
	}
}

func initConfig() {
	viper.SetConfigType("toml")
	viper.SetConfigFile("config.toml")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}

	logrus.Info("Using Config file: ", viper.ConfigFileUsed())

	if viper.GetBool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Warn("Comment service is Running in Debug Mode")
		return
	}
	logrus.SetLevel(logrus.InfoLevel)
	logrus.Warn("Comment service is Running in Production Mode")
	logrus.SetFormatter(&logrus.JSONFormatter{})
}

func init() {
	initConfig()
}
