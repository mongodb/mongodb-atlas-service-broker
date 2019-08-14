package broker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCatalog(t *testing.T) {
	broker, _ := setupTest()

	services, err := broker.Services(context.Background())

	assert.NoError(t, err)
	assert.NotZero(t, len(services), "Expected a non-zero amount of services")

	for _, service := range services {
		assert.NotZerof(t, len(service.Plans), "Expected a non-zero amount of plans for service %s", service.Name)
	}
}
