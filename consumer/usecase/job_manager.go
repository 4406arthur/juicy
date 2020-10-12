package usecase

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/4406arthur/juicy/consumer/alert"
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
	finished   chan<- bool
	alert      alert.Alert
}

//NewJobManager ...
func NewJobManager(poolSize int, validate *validator.Validate, httpCli *httpclient.Client, jobQueue <-chan *nats.Msg, ansCh chan<- []byte, finished chan<- bool, alert alert.Alert) domain.JobManager {
	wp := workerpool.New(poolSize)
	return &jobManager{
		workerPool: wp,
		validate:   validate,
		httpClient: httpCli,
		jobQueue:   jobQueue,
		ansCh:      ansCh,
		finished:   finished,
		alert:      alert,
	}
}

var (
	opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "juicy_manager_total_jobs",
		Help: "The total number of job",
	})
)

func (m *jobManager) Start(ctx context.Context) {
	log.Printf("[Info] job manager starting\n")
	for {
		select {
		case element := <-m.jobQueue:
			var job domain.Job
			ffjson.Unmarshal(element.Data, &job)
			err := m.validate.Struct(job)
			if err != nil {
				log.Printf("[Error] got wrong job format: %s", err.Error())
				continue
			}
			log.Printf("[Debug] Recevie job: %s", job.QuesionID)
			m.workerPool.Submit(
				func() {
					opsProcessed.Inc()
					m.Task(job)
				})
		case <-ctx.Done():
			log.Println("[Info] close workers")
			m.Stop()
			return
		}
	}
}

func (m *jobManager) Stop() {
	m.workerPool.StopWait()
	log.Printf("[Info] already completed pending task")
	m.finished <- true
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
		errMsg := fmt.Sprintf("[ERROR] Qid: %s TeamID: %s Got error with: %s \n", job.QuesionID, job.TeamID, err.Error())
		log.Printf(errMsg)
		m.alert.PushNotify(errMsg)
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
		errMsg := fmt.Sprintf("[ERROR] Qid: %s TeamID: %s Got error when validate inference server resp: %s \n", job.QuesionID, job.TeamID, err.Error())
		log.Printf(errMsg)
		m.alert.PushNotify(errMsg)
		jsonByte, _ := ffjson.Marshal(&domain.Respond{
			TeamID:    job.TeamID,
			QuesionID: job.QuesionID,
			ErrorMsg:  err.Error(),
		})
		opsError.Inc()
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
		//log.Printf("[Error] http invoke error: %s \n", err.Error())
		return respond, err
	}
	if resp.StatusCode >= 400 {
		//log.Printf("[Error] http invoke got wrong statusCode: %d \n", resp.StatusCode)
		return respond, errors.New("wrong http status code: " + strconv.Itoa(resp.StatusCode))
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//log.Printf("[Error] parse respond error: %s \n", err.Error())
		return respond, err
	}

	ffjson.Unmarshal(respBody, &respond)
	return respond, nil
}

func alertPush(msg string) {
	curlCMD := fmt.Sprintf("curl -X POST -H 'Content-type: application/json' --data '{\"text\": \"%s\"}' https://hooks.slack.com/services/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", msg)
	cmdList := strings.Split(curlCMD, " ")
	cmd := exec.Command(cmdList[0], cmdList[1:]...)
	err := cmd.Run()
	if err != nil {
		log.Printf("curl error: %s \n", err.Error())
	}
}
