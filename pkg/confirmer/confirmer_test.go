package confirmer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfirmerYes(t *testing.T) {
	c := New()

	req := c.NewRequest(1 * time.Second)

	go func() {
		assert.NoError(t, c.SetResponse(req.ID(), true))
	}()

	res, err := req.Result()
	assert.NoError(t, err)
	assert.True(t, res)
}

func TestConfirmerNo(t *testing.T) {
	c := New()

	req := c.NewRequest(1 * time.Second)

	go func() {
		assert.NoError(t, c.SetResponse(req.ID(), false))
	}()

	res, err := req.Result()
	assert.NoError(t, err)
	assert.False(t, res)
}

func TestConfirmerTimeout(t *testing.T) {
	c := New()

	req := c.NewRequest(1 * time.Second)

	go func() {
		time.Sleep(2 * time.Second)
		assert.NoError(t, c.SetResponse(req.ID(), true))
	}()

	_, err := req.Result()
	assert.Error(t, err)
}
