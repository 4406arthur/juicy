package usecase

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/4406arthur/juicy/domain"
	"github.com/gammazero/workerpool"
	"github.com/gojektech/heimdall/v6/httpclient"
	"github.com/nats-io/nats.go"
	"github.com/pquerna/ffjson/ffjson"
)

type workerUsecase struct {
	workerPool *workerpool.WorkerPool
	httpClient *httpclient.Client
	jobQueue   <-chan *nats.Msg
	//jobRepo  domain.JobRepository
}

//NewWorkerUsecase ...
func NewWorkerUsecase(poolSize int, httpCli *httpclient.Client, jobQueue <-chan *nats.Msg) domain.WorkerUsecase {
	wp := workerpool.New(poolSize)
	return &workerUsecase{
		workerPool: wp,
		httpClient: httpCli,
		jobQueue:   jobQueue,
	}
}

func (w *workerUsecase) Start() {
	var job domain.Job
	for element := range w.jobQueue {
		fmt.Printf("receive job bytes: %s", string(element.Data))
		err := ffjson.Unmarshal(element.Data, &job)
		if err != nil {
			fmt.Printf("job unserialize failed: %s", err.Error())
			continue
		}
		w.workerPool.Submit(func() {
			w.AssignmentHandler(job)
		})
	}
}

func (w *workerUsecase) AssignmentHandler(rq domain.Job) (domain.Respond, error) {
	jsonByte, _ := ffjson.Marshal(rq)
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	resp, err := w.httpClient.Post(
		rq.ServerEndpoint,
		bytes.NewBuffer(jsonByte),
		headers,
	)
	if err != nil {
		fmt.Printf("http invoke error: %s ", err.Error())
	}
	respBody, _ := ioutil.ReadAll(resp.Body)
	var respond domain.Respond
	ffjson.Unmarshal(respBody, &respond)

	return respond, nil
}

// func (w *workerUsecase) ScheduledAssignmentHandler(endpoint string, rq domain.Job) (domain.Respond, error) {
// 	var respond domain.Respond
// 	return respond, nil
// }
