package health

import (
	"testing"
)

func TestCheckRabbitMQHealth(t *testing.T) {

	r := NewRabbitMQHealth()

	healthy, _ := r.CheckRabbitMQHealth()
	if !healthy {
		t.Fatal("rabbimq is not healthy")
	}

}
