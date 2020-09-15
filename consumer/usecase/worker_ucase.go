package usecase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/4406arthur/juicy/domain"
	"github.com/gammazero/workerpool"
	"github.com/gojektech/heimdall/httpclient"
)

type workerUsecase struct {
	workerPool *workerpool.WorkerPool
	httpClient *httpclient.Client
	//jobRepo  domain.JobRepository
}

//NewWorkerUsecase ...
func NewWorkerUsecase(poolSize int, httpCli *httpclient.Client) domain.WorkerUsecase {
	wp := workerpool.New(poolSize)
	return &workerUsecase{
		workerPool: wp,
		httpClient: httpCli,
	}
}

func (w *workerUsecase) AssignmentHandler(endpoint string, rq domain.Request) (domain.Respond, error) {
	jsonByte, _ := json.Marshal(rq)
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	resp, err := w.httpClient.Post(
		endpoint,
		bytes.NewBuffer(jsonByte),
		headers,
	)
	if err != nil {
		fmt.Printf("http invoke error: %s ", err.Error())
	}
	respBody, _ := ioutil.ReadAll(resp.Body)
	var respond domain.Respond
	json.Unmarshal(respBody, &respond)

	return respond, nil
}

func (w *workerUsecase) ScheduledAssignmentHandler(endpoint string, rq domain.Request) (domain.Respond, error) {
	var respond domain.Respond
	return respond, nil
}
