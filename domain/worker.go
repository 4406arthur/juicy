package domain

import (
	"context"
)

//Request ...
type Request struct {
	ScoringTeamID int    `json:"scoring_system_team_id"`
	QuesionID     int    `json:"qid" validate:"required"`
	EsunUUID      string `json:"esun_uuid"`
	ServerUUID    string `json:"server_uuid"`
	EsunTimestamp int64  `json:"esun_timestamp"`
	News          string `json:"news"`
	Retry         int    `json:"retry"`
}

//Respond ...
type Respond struct {
	ClientID        string   `json:"client_id"`
	QuesionID       int      `json:"quesion_id"`
	EsunUUID        string   `json:"esun_uuid,omitempty" validate:"required"`
	ServerUUID      string   `json:"server_uuid,omitempty" validate:"required"`
	ServerTimestamp string   `json:"server_timestamp,omitempty" validate:"required"`
	Answer          []string `json:"answer" validate:"required"`
	ErrorMsg        string   `json:"error_msg,omitempty"`
}

//Job ...
type Job struct {
	ClientID       string  `json:"client_id" validate:"required"`
	QuesionID      int     `json:"qid" validate:"required"`
	ServerEndpoint string  `json:"server_endpoint" validate:"required"`
	Payload        Request `json:"payload" validate:"required"`
}

//JobManager ...
type JobManager interface {
	Start(context.Context)
	PostInferenceHandler(string, Request) (Respond, error)
	Stop()
	// ScheduledAssignmentHandler(endpoint string, rq Job) (Respond, error)
}
