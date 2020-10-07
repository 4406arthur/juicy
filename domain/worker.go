package domain

import (
	"context"
)

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
	TeamID          string   `json:"scoring_system_team_id" `
	QuesionID       string   `json:"quesion_id" `
	EsunUUID        string   `json:"esun_uuid,omitempty" validate:"required"`
	ServerUUID      string   `json:"server_uuid,omitempty" validate:"required"`
	ServerTimestamp int      `json:"server_timestamp,omitempty" validate:"required"`
	Answer          []string `json:"answer,omitempty" validate:"required"`
	ErrorMsg        string   `json:"error_msg,omitempty"`
}

//Job ...
type Job struct {
	TeamID         string  `json:"scoring_system_team_id" validate:"required"`
	QuesionID      string  `json:"quesion_id" validate:"required"`
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
