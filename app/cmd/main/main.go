package main

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-chi/chi/v5"
	"orderInfo/app/internal/services/orderInfo"
	"orderInfo/app/internal/storage/postgreSQL"
	"orderInfo/app/internal/transport/handlers"
	consumer "orderInfo/app/internal/transport/kafka"
	"orderInfo/app/pkg/logger"
	"time"
)

const (
	broker = "kafka:29092"
	topic  = "createMess"
)

func main() {
	time.Sleep(time.Second * 4)
	db, err := postgreSQL.OpenDB(context.Background(), "postgres://postgres:admin@postgresql:5432/wb_orderInfo_database?sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.DB.Close(context.Background())
	log := logger.SetupLogger("dev")

	service := orderInfo.New(log, db, db, db)

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": broker,
		"group.id":          "1",
		"auto.offset.reset": "smallest",
	})
	if err != nil {
		panic(err)
	}

	orderConsumer := consumer.NewOrderConsumer(c, topic, service, log)
	go orderConsumer.Init()

	s := handlers.NewServer(service, service)

	r := chi.NewRouter()
	fmt.Println("All is fine!")
	s.Start(r, ":8080")
}
