package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	// _ "github.com/lib/pq"

	_natsDeliver "github.com/4406arthur/juicy/consumer/delivery/nats"
	_workerUsecase "github.com/4406arthur/juicy/consumer/usecase"
	"github.com/gojektech/heimdall"
	"github.com/gojektech/heimdall/v6/httpclient"
	"github.com/nats-io/nats.go"
	"github.com/spf13/viper"
)

var usageStr = `
AI Customer service controller

Server Options:
    -c, --config <file>              Configuration file path
    -h, --help                       Show this message
    -v, --version                    Show version
`

// usage will print out the flag options for the server.
func usage() {
	fmt.Printf("%s\n", usageStr)
	os.Exit(0)
}

func setup(path string) *viper.Viper {
	v := viper.New()
	v.SetConfigType("json")
	v.SetConfigName("config")
	if path != "" {
		v.AddConfigPath(path)
	} else {
		v.AddConfigPath("./config/")
	}

	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
	return v
}

var version string

func printVersion() {
	fmt.Printf(`Juicy worker %s, Compiler: %s %s, Copyright (C) 2020 EsunBank, Inc.`,
		version,
		runtime.Compiler,
		runtime.Version())
	fmt.Println()
}

func main() {

	var configFile string
	var showVersion bool
	version = "0.0.1"
	flag.BoolVar(&showVersion, "v", false, "Print version information.")
	flag.StringVar(&configFile, "c", "", "Configuration file path.")
	flag.Usage = usage
	flag.Parse()

	if showVersion {
		printVersion()
		os.Exit(0)
	}

	viperConfig := setup(configFile)
	// dbHost := viperConfig.GetString(`database.host`)
	// dbPort := viperConfig.GetUint(`database.port`)
	// dbUser := viperConfig.GetString(`database.user`)
	// dbPass := viperConfig.GetString(`database.pass`)
	// dbName := viperConfig.GetString(`database.name`)
	// psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
	// 	"password=%s dbname=%s sslmode=disable",
	// 	dbHost, dbPort, dbUser, dbPass, dbName)
	// dbConn, err := sql.Open("postgres", psqlInfo)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer func() {
	// 	err := dbConn.Close()
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// }()

	natsHost := viperConfig.GetString(`nats.host`)
	messageQueue := _natsDeliver.NewMessageQueue(natsHost)

	poolSize := viperConfig.GetInt(`worker.poolSize`)

	//define a retry http cli
	timeout := 3000 * time.Millisecond
	httpCli := httpclient.NewClient(
		httpclient.WithHTTPTimeout(timeout),
		httpclient.WithRetryCount(2),
		httpclient.WithRetrier(heimdall.NewRetrier(heimdall.NewConstantBackoff(10*time.Millisecond, 50*time.Millisecond))),
	)

	//could be buffer queue
	jobQueue := make(chan *nats.Msg)
	worker := _workerUsecase.NewWorkerUsecase(poolSize, httpCli, jobQueue)
	messageQueue.Subscribe("news", "juicy-workers", jobQueue)
	worker.Start()

}
