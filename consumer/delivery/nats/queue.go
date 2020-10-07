package nats

import (
	"log"

	"github.com/nats-io/nats.go"
)

//MessageQueue ...
type MessageQueue struct {
	natConn *nats.Conn
	//injection usecase like how to handle incoming message
}

//NewMessageQueue ...
func NewMessageQueue(conn string) *MessageQueue {
	//TODO support opts flag
	opts := []nats.Option{nats.Name("juicy"), nats.ErrorHandler(logSlowConsumer)}
	nc, err := nats.Connect(conn, opts...)
	if err != nil {
		log.Fatalf("Failed to connect Nats: %s", err.Error())
	}
	return &MessageQueue{
		natConn: nc,
	}
}

//Subscribe ...
func (s *MessageQueue) Subscribe(subj, queueName string, ch chan *nats.Msg) {
	s.natConn.QueueSubscribeSyncWithChan(subj, queueName, ch)
}

//Publish ...
func (s *MessageQueue) Publish(subj string, ch chan []byte) {
	for msg := range ch {
		s.natConn.Publish(subj, msg)
	}
}

func logSlowConsumer(nc *nats.Conn, sub *nats.Subscription, err error) {
	if err == nats.ErrSlowConsumer {
		dropped, _ := sub.Dropped()
		log.Printf("Slow consumer on subject %q: dropped %d messages\n",
			sub.Subject, dropped)
	}
}
