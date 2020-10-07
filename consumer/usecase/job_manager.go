package usecase

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/4406arthur/juicy/domain"
	"github.com/gammazero/workerpool"
	"github.com/go-playground/validator/v10"
	"github.com/gojektech/heimdall/v6/httpclient"
	"github.com/nats-io/nats.go"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type jobManager struct {
	workerPool *workerpool.WorkerPool
	validate   *validator.Validate
	httpClient *httpclient.Client
	jobQueue   <-chan *nats.Msg
	ansCh      chan<- []byte
	//jobRepo  domain.JobRepository
}

//NewJobManager ...
func NewJobManager(poolSize int, validate *validator.Validate, httpCli *httpclient.Client, jobQueue <-chan *nats.Msg, ansCh chan<- []byte) domain.JobManager {
	wp := workerpool.New(poolSize)
	return &jobManager{
		workerPool: wp,
		validate:   validate,
		httpClient: httpCli,
		jobQueue:   jobQueue,
		ansCh:      ansCh,
	}
}

var (
	opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "juicy_manager_total_jobs",
		Help: "The total number of job",
	})
)

func (m *jobManager) Start(ctx context.Context) {
	for {
		select {
		case element := <-m.jobQueue:
			var job domain.Job
			ffjson.Unmarshal(element.Data, job)
			err := m.validate.Struct(job)
			if err != nil {
				log.Printf("got wrong job format: %s", err.Error())
				continue
			}
			log.Printf("Recevie job: %s", job.QuesionID)
			m.workerPool.Submit(
				func() {
					opsProcessed.Inc()
					m.Task(job)
				})
		case <-ctx.Done():
			log.Println("close workers")
			m.Stop()
			return
		}
	}
}

func (m *jobManager) Stop() {
	m.workerPool.StopWait()
	// close(m.jobQueue)
	close(m.ansCh)
}

var (
	opsError = promauto.NewCounter(prometheus.CounterOpts{
		Name: "juicy_manager_failed_job",
		Help: "The failed number of job",
	})
)

func (m *jobManager) Task(job domain.Job) {
	var Respond domain.Respond
	Respond, err := m.PostInferenceHandler(job.ServerEndpoint, job.Payload)
	if err != nil {
		log.Printf("Got error with: %s \n", err.Error())
		jsonByte, _ := ffjson.Marshal(&domain.Respond{
			TeamID:    job.TeamID,
			QuesionID: job.QuesionID,
			ErrorMsg:  err.Error(),
		})
		opsError.Inc()
		m.ansCh <- jsonByte
		return
	}
	err = m.validate.Struct(&Respond)
	if err != nil {
		log.Printf("Got error when validate inference server resp: %s \n", err.Error())
		jsonByte, _ := ffjson.Marshal(&domain.Respond{
			TeamID:    job.TeamID,
			QuesionID: job.QuesionID,
			ErrorMsg:  err.Error(),
		})
		m.ansCh <- jsonByte
		return
	}
	Respond.TeamID = job.TeamID
	Respond.QuesionID = job.QuesionID
	jsonByte, _ := ffjson.Marshal(Respond)
	m.ansCh <- jsonByte
}

//PostInferenceHandler ...
func (m *jobManager) PostInferenceHandler(endpoint string, rq domain.Request) (domain.Respond, error) {
	jsonByte, _ := ffjson.Marshal(rq)
	var respond domain.Respond
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	resp, err := m.httpClient.Post(
		endpoint,
		bytes.NewBuffer(jsonByte),
		headers,
	)
	if err != nil {
		log.Printf("http invoke error: %s \n", err.Error())
		return respond, err
	}
	if resp.StatusCode >= 400 {
		log.Printf("http invoke got wrong statusCode: %d \n", resp.StatusCode)
		return respond, errors.New("wrong http status code")
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("parse respond error: %s \n", err.Error())
		return respond, err
	}

	ffjson.Unmarshal(respBody, &respond)
	return respond, nil
}
