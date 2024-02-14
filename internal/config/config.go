package config

import (
	"errors"
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
	MPurl = os.Getenv("MP_URL")
	if MPurl == "" {
		return errors.New("MP_URL not found")
	}
	RabbitmqURL = os.Getenv("RABBITMQ_URL")
	if RabbitmqURL == "" {
		return errors.New("RABBITMQ_URL not found")
	}
	MpToken = os.Getenv("MP_TOKEN")
	if MpToken == "" {
		return errors.New("MP_TOKEN not found")
	}
	MpUrlCCToken = os.Getenv("MP_URL_CREDIT_CARD")
	if MpUrlCCToken == "" {
		return errors.New("MP_URL_CREDIT_CARD not found")
	}	
	return nil
}
