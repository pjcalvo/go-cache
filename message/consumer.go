package message

import (
	"fmt"
	"time"

	"github.com/pjcalvo/go-cache/utils"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaConsumer struct {
	*kafka.Consumer
}

func InitConsumer() (*KafkaConsumer, error) {
	// set config properties
	conf := utils.ReadConfig(configFile)
	conf["group.id"] = "kafka-go-getting-started"
	conf["auto.offset.reset"] = "earliest"

	c, err := kafka.NewConsumer(&conf)

	if err != nil {
		return nil, fmt.Errorf("Failed to create consumer: %s", err)
	}

	err = c.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to subscribe to topic: %s", err)
	}

	return &KafkaConsumer{c}, nil

}

func (c *KafkaConsumer) Read() {
	ev, err := c.ReadMessage(100 * time.Millisecond)
	if err != nil {
		// Errors are informational and automatically handled by the consumer
		return
	}
	fmt.Printf("one time cache requested from db :%s\n", ev.Key)
}
