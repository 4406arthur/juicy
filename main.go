package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	// _ "github.com/lib/pq"
	"github.com/4406arthur/juicy/consumer/alert"
	_natsDeliver "github.com/4406arthur/juicy/consumer/delivery/nats"
	_jobManager "github.com/4406arthur/juicy/consumer/usecase"
	"github.com/DataDog/datadog-go/statsd"
	"github.com/go-playground/validator/v10"
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

// use a single instance of Validate, it caches struct info
var validate *validator.Validate

func setLogPath(filePath string) {
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
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
	logPath := viperConfig.GetString(`log.path`)
	if logPath != "" {
		setLogPath(logPath)
	}

	natsHost := viperConfig.GetString(`nats.host`)
	messageQueue := _natsDeliver.NewMessageQueue(natsHost)
	jobRetry := viperConfig.GetInt(`job.retryCount`)
	timeout := viperConfig.GetInt(`job.timeout`)
	//define a retry http cli
	httpCli := httpclient.NewClient(
		httpclient.WithHTTPTimeout(time.Duration(timeout)*time.Millisecond),
		httpclient.WithRetryCount(jobRetry),
		httpclient.WithRetrier(heimdall.NewRetrier(heimdall.NewConstantBackoff(10*time.Millisecond, 50*time.Millisecond))),
	)

	//setup slack alert
	slackAlert := alert.NewSlackWebhook(httpCli, viperConfig.GetString(`slackAlert.endpoint`))

	//setup a datadog metrics endpoint
	statsd, _ := statsd.New("127.0.0.1:8125")

	//could be buffer queue
	jobQueue := make(chan *nats.Msg, 2000)
	defer close(jobQueue)
	ansCh := make(chan []byte)
	defer close(ansCh)
	// used to catch os signal
	// syscall.SIGINT and syscall.SIGTERM
	finished := make(chan bool)
	defer close(finished)
	ctx := withContextFunc(context.Background(), func() {
		log.Println("[Info] cancel from ctrl+c event")
	})

	sub := viperConfig.GetString(`nats.sub`)
	pub := viperConfig.GetString(`nats.pub`)
	messageQueue.Subscribe(sub, "juicy-workers", jobQueue)
	validate = validator.New()
	wokerPoolSize := viperConfig.GetInt(`worker.poolSize`)
	jobManager := _jobManager.NewJobManager(wokerPoolSize, validate, httpCli, jobQueue, ansCh, finished, slackAlert, statsd)
	go jobManager.Start(ctx)
	go messageQueue.Publish(pub, ansCh)

	<-finished
}

func withContextFunc(ctx context.Context, f func()) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(c)

		select {
		case <-ctx.Done():
		case <-c:
			cancel()
			f()
		}
	}()

	return ctx
}
