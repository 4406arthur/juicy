package domain

//Request ...
type Request struct {
	EsunUUID      string `json:"esun_uuid"`
	ServerUUID    string `json:"server_uuid"`
	EsunTimestamp int    `json:"esun_timestamp"`
	News          string `json:"news"`
	Retry         int    `json:"retry"`
}

//Respond ...
type Respond struct {
	EsunUUID        string   `json:"esun_uuid"`
	ServerUUID      string   `json:"server_uuid"`
	ServerTimestamp int      `json:"server_timestamp"`
	Answer          []string `json:"answer"`
}

//WorkerUsecase ...
type WorkerUsecase interface {
	AssignmentHandler(endpoint string, rq Request) (Respond, error)
	ScheduledAssignmentHandler(endpoint string, rq Request) (Respond, error)
}
