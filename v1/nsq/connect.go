package nsq

import (
	"log"
	"os"

	"github.com/nsqio/go-nsq"
)

type nopLogger struct{}
var producer *nsq.Producer

func (*nopLogger) Output(int, string) error {
	return nil
}

func Start() (*nsq.Producer, error) {
	if producer == nil || producer.Ping() != nil {
		config := nsq.NewConfig()
		config.Set("heartbeat_interval", "10s")

		var err error
		producer, err = nsq.NewProducer(os.Getenv("NSQD_HOST"), config)
		if err != nil {
			log.Println("failed to connect to nsqd: ", err.Error())
			return nil, err
		}
	
		producer.SetLogger(&nopLogger{}, 0)
	}
	return producer, nil
}