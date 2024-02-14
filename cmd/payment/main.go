package main

import (
	"log"
	"payment/internal/config"
	"payment/internal/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	if err := config.LoadConfig(); err != nil {
		log.Fatal(err)
	}
	conn, err := amqp.Dial(config.RabbitmqURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	rabbitmqService := rabbitmq.NewService(conn)
	

	if err = rabbitmqService.ConsumePayment(); err != nil {
		log.Fatal(err)
	}
}
