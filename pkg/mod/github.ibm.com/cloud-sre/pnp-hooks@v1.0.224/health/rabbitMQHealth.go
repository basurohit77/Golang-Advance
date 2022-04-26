package health

import (
	"github.ibm.com/cloud-sre/pnp-hooks/shared"
	"time"
)

// RabbitMQHealth data
type RabbitMQHealth struct {
	isHealthy bool
	healthDescription string
	intervalToCheckInSeconds int
}

// NewRabbitMQHealth - constructor for RabbitMQ health checker
func NewRabbitMQHealth() *RabbitMQHealth {
	rabbitMQHealth := &RabbitMQHealth{
		isHealthy: true,
		healthDescription: "",
		intervalToCheckInSeconds: 60,
	}
	go monitorRabbitMQ(rabbitMQHealth)
	return rabbitMQHealth
}

// CheckRabbitMQHealth - Get latest RabbitMQ health check
func (rabbitMQHealth RabbitMQHealth) CheckRabbitMQHealth() (bool, string) {
	return rabbitMQHealth.isHealthy, rabbitMQHealth.healthDescription
}

// Continually check the health of RabbitMQ and update the latest status
func monitorRabbitMQ(rabbitMQHealth *RabbitMQHealth) {
	for {
		rabbitMQHealth.isHealthy = shared.TestConnectionToRabbitMQ()
		time.Sleep(time.Duration(rabbitMQHealth.intervalToCheckInSeconds) * time.Second)
	}
}
