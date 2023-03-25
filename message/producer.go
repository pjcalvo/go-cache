package message

import (
	"fmt"

	"github.com/pjcalvo/go-cache/utils"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaProducer struct {
	*kafka.Producer
}

func InitProducer() (*KafkaProducer, error) {
	conf := utils.ReadConfig(configFile)

	p, err := kafka.NewProducer(&conf)
	if err != nil {
		return nil, fmt.Errorf("Failed to create producer: %s", err)
	}

	// Go-routine to handle message delivery reports and
	// possibly other event types (errors, stats, etc)
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Failed to deliver message: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Produced event to topic %s: key = %-10s value = %s\n",
						*ev.TopicPartition.Topic, string(ev.Key), string(ev.Value))
				}
			}
		}
	}()

	return &KafkaProducer{p}, nil
}

func (p *KafkaProducer) Send(m Message) error {
	top := fmt.Sprintf(topic)
	err := p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &top, Partition: kafka.PartitionAny},
		Key:            []byte(m.MessageID),
	}, nil)
	if err != nil {
		return fmt.Errorf("error producing messages: %w", err)
	}

	// Wait for all messages to be delivered
	p.Flush(15 * 1000)
	return nil
}
