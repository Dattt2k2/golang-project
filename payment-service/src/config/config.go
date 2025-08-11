package config

type Config struct {
	KafkaBroker       string `json:"kafka_broker"`
	PaymentGatewayURL string `json:"payment_gateway_url"`
	PaymentGatewayKey string `json:"payment_gateway_key"`
	PaymentGatewaySecret string `json:"payment_gateway_secret"`
}