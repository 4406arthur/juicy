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

func (w *jobManager) Start(ctx context.Context) {
	var job domain.Job
	for {
		select {
		case element := <-w.jobQueue:
			ffjson.Unmarshal(element.Data, &job)
			err := w.validate.Struct(&job)
			if err != nil {
				log.Printf("got wrong job format: %s", err.Error())
				continue
			}
			w.workerPool.Submit(
				func() {
					w.Task(&job)
				})
		case <-ctx.Done():
			log.Println("close workers")
			w.Stop()
			return
		}
	}
}

func (w *jobManager) Stop() {
	w.workerPool.StopWait()
}

func (w *jobManager) Task(job *domain.Job) {
	var Respond domain.Respond
	Respond, err := w.PostInferenceHandler(job.ServerEndpoint, job.Payload)
	if err != nil {
		log.Printf("Got error with: %s \n", err.Error())
		jsonByte, _ := ffjson.Marshal(&domain.Respond{
			TeamID:    job.TeamID,
			QuesionID: job.QuesionID,
			ErrorMsg:  err.Error(),
		})
		w.ansCh <- jsonByte
		return
	}
	err = w.validate.Struct(&Respond)
	if err != nil {
		log.Printf("Got error when validate inference server resp: %s \n", err.Error())
		jsonByte, _ := ffjson.Marshal(&domain.Respond{
			TeamID:    job.TeamID,
			QuesionID: job.QuesionID,
			ErrorMsg:  err.Error(),
		})
		w.ansCh <- jsonByte
		return
	}
	Respond.TeamID = job.TeamID
	Respond.QuesionID = job.QuesionID
	jsonByte, _ := ffjson.Marshal(Respond)
	w.ansCh <- jsonByte
}

//PostInferenceHandler ...
func (w *jobManager) PostInferenceHandler(endpoint string, rq domain.Request) (domain.Respond, error) {
	jsonByte, _ := ffjson.Marshal(rq)
	var respond domain.Respond
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	resp, err := w.httpClient.Post(
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
