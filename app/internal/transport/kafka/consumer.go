// Kafka consumer
package consumer

import (
	"context"
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log/slog"
	"orderInfo/app/internal/models"
)

type MessageAdder interface {
	AddOrder(ctx context.Context, order models.Order) error
}

type OrderConsumer struct {
	consumer *kafka.Consumer
	topic    string
	adder    MessageAdder
	log      *slog.Logger
}

func NewOrderConsumer(c *kafka.Consumer, topic string, adder MessageAdder, log *slog.Logger) *OrderConsumer {
	return &OrderConsumer{
		consumer: c,
		topic:    topic,
		adder:    adder,
		log:      log,
	}
}

func (oc *OrderConsumer) Init() {
	go func() {
		for {
			oc.messListener()
		}
	}()
}

func (oc *OrderConsumer) messListener() {
	err := oc.consumer.Subscribe(oc.topic, nil)
	if err != nil {
		oc.log.Error("Consumer Error", slog.String("error", err.Error()))
		return
	}

	for {
		ev := oc.consumer.Poll(100)
		switch e := ev.(type) {
		case *kafka.Message:
			oc.log.Info("New msg revived")
			buf := e.Value
			order := models.Order{}

			err = json.Unmarshal(buf, &order)
			if err != nil {
				oc.log.Error("Failed to unmarshal received msg")
				return
			}

			//fmt.Println("Revived:", order)

			err := oc.adder.AddOrder(context.Background(), order)
			if err != nil {
				oc.log.Error("Consumer error", slog.String("error", err.Error()))
				return
			}
		case *kafka.Error:
			oc.log.Error("Consumer Error", slog.String("error", err.Error()))
			return
		}
	}
}
