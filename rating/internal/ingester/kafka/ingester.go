package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/tj330/bookapp/rating/pkg/model"
)

type Ingester struct {
	consumer *kafka.Consumer
	topic    string
}

func NewIngester(addr string, groupID string, topic string) (*Ingester, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": addr,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, err
	}

	return &Ingester{consumer, topic}, nil
}

func (i *Ingester) Ingest(ctx context.Context) (chan model.RatingEvent, error) {
	fmt.Println("Starting kafka Ingester")
	if err := i.consumer.SubscribeTopics([]string{i.topic}, nil); err != nil {
		return nil, err
	}
	ch := make(chan model.RatingEvent, 1)

	go func() {
		defer close(ch)
		defer i.consumer.Close()

		for {
			msg, err := i.consumer.ReadMessage(100)
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					continue
				}
			}

			var event model.RatingEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				fmt.Println("Unmarshal error:", err)
				continue
			}

			select {
			case ch <- event:
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch, nil
}
