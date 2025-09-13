package config

import "os"

type Config struct {
	KafkaBroker       string `json:"kafka_broker"`
	KafkaTopic        string `json:"kafka_topic"`
	PaymentGatewayURL string `json:"payment_gateway_url"`
	PaymentGatewayKey string `json:"payment_gateway_key"`
	PaymentGatewaySecret string `json:"payment_gateway_secret"`
}

func LoadConfig() (*Config, error) {
	return &Config{
		KafkaBroker:       os.Getenv("KAFKA_BROKER"),
		KafkaTopic:        os.Getenv("KAFKA_TOPIC"),
		PaymentGatewayURL: os.Getenv("PAYMENT_GATEWAY_URL"),
		PaymentGatewayKey: os.Getenv("PAYMENT_GATEWAY_KEY"),
		PaymentGatewaySecret: os.Getenv("PAYMENT_GATEWAY_SECRET"),
	}, nil
}