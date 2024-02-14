package config

import (

	"os"
)

var (
	MPurl string
	RabbitmqURL string
	MpToken string
	MpUrlCCToken string

	// MPurl = "https://api.mercadopago.com/v1/payments"
	// RabbitmqURL = "amqp://guest:guest@localhost:5672/"
)

func LoadConfig() error {
	MPurl = "https://api.mercadopago.com/v1/payments"
	
	RabbitmqURL = os.Getenv("RABBITMQ_URL")
	
	MpToken = "APP_USR-2263346854076132-011400-800633c8a8e6697eff787a71b9a4dade-1628717800"

	MpUrlCCToken = "https://api.mercadopago.com/v1/card_tokens"
	
	return nil
}
