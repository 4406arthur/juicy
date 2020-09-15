package delivery

import (
	"log"

	"github.com/nats-io/nats.go"
)

//Subscriber ...
type Subscriber struct {
	natConn *nats.Conn
	//injection usecase like how to handle incoming message
}

//NewSubscriber ...
func NewSubscriber(conn string) *Subscriber {
	//TODO support opts flag
	opts := []nats.Option{nats.Name("juicy")}
	nc, err := nats.Connect(conn, opts...)
	if err != nil {
		log.Fatalf("Failed to connect Nats: %s", err.Error())
	}
	return &Subscriber{
		natConn: nc,
	}
}

//Subscribe ...
func (s *Subscriber) Subscribe(subj, queueName string, ch chan *nats.Msg) {
	s.natConn.QueueSubscribeSyncWithChan(subj, queueName, ch)
}
