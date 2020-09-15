package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_natsDeliver "github.com/4406arthur/juicy/delivery/nats"
	_workerUsecase "github.com/4406arthur/juicy/usecase"
	"github.com/gojektech/heimdall"
	"github.com/gojektech/heimdall/v6/httpclient"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigFile(`config.json`)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	if viper.GetBool(`debug`) {
		log.Println("Service RUN on DEBUG mode")
	}
}

func main() {
	dbHost := viper.GetString(`database.host`)
	dbPort := viper.GetUint(`database.port`)
	dbUser := viper.GetString(`database.user`)
	dbPass := viper.GetString(`database.pass`)
	dbName := viper.GetString(`database.name`)
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)
	dbConn, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		err := dbConn.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	natsHost := viper.GetString(`nats.host`)
	_natsDeliver.NewSubscriber(natsHost)

	poolSize := viper.GetInt(`worker.poolSize`)

	// First set a backoff mechanism. Constant backoff increases the backoff at a constant rate
	backoffInterval := 2 * time.Millisecond
	// Define a maximum jitter interval. It must be more than 1*time.Millisecond
	maximumJitterInterval := 5 * time.Millisecond
	backoff := heimdall.NewConstantBackoff(backoffInterval, maximumJitterInterval)
	// Create a new retry mechanism with the backoff
	retrier := heimdall.NewRetrier(backoff)
	timeout := 5000 * time.Millisecond
	// Create a new client, sets the retry mechanism, and the number of times you would like to retry
	client := httpclient.NewClient(
		httpclient.WithHTTPTimeout(timeout),
		httpclient.WithRetrier(retrier),
		httpclient.WithRetryCount(3),
	)
	worker := _workerUsecase.NewWorkerUsecase(poolSize, client)

}
