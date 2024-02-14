package rabbitmq

import (
	"encoding/json"
	"log"
	"payment/internal/model"
	"payment/internal/payment"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Service interface {
	ConsumePayment() error
}

type service struct {
	conn *amqp.Connection
}

func NewService(conn *amqp.Connection) Service {
	return &service{conn}
}

func (s *service) processPayment(paymentData *model.CreatePayment) {
	paymentService := payment.NewPaymentService(s.conn)
	paymentService.CreatePayment(paymentData)
}

func (s *service) ConsumePayment() error {
	channel, err := s.conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	queueName := "payments"
	msgs, err := channel.Consume(
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	log.Println("Consumer online, awaiting messages")
	for msg := range msgs {
		var paymentData model.CreatePayment
		err := json.Unmarshal(msg.Body, &paymentData)
		if err != nil {
			log.Fatalf("Error to decode message: %s\n", err.Error())
			continue
		}
		s.processPayment(&paymentData)
	}
	return nil
}

