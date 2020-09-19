package domain

//Request ...
// type Request struct {
// 	EsunUUID      string `json:"esun_uuid"`
// 	ServerUUID    string `json:"server_uuid"`
// 	EsunTimestamp int    `json:"esun_timestamp"`
// 	News          string `json:"news"`
// 	Retry         int    `json:"retry"`
// }

//Respond ...
type Respond struct {
	EsunUUID        string   `json:"esun_uuid"`
	ServerUUID      string   `json:"server_uuid"`
	ServerTimestamp int      `json:"server_timestamp"`
	Answer          []string `json:"answer"`
}

//Job ...
type Job struct {
	ServerEndpoint string `json:"server_endpoint"`
	EsunUUID       string `json:"esun_uuid"`
	ServerUUID     string `json:"server_uuid"`
	EsunTimestamp  int    `json:"esun_timestamp"`
	News           string `json:"news"`
	Retry          int    `json:"retry"`
}

//WorkerUsecase ...
type WorkerUsecase interface {
	Start()
	AssignmentHandler(rq Job) (Respond, error)
	// ScheduledAssignmentHandler(endpoint string, rq Job) (Respond, error)
}
