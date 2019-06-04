package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/threefoldtech/zbus"
	"github.com/threefoldtech/zbus/examples/server/stubs"
)

func TestAdd(t *testing.T) {
	client, err := zbus.NewRedisClient("tcp://localhost:6379")
	if err != nil {
		panic(err)
	}

	calculator := stubs.NewCalculatorStub(client)

	result := calculator.Add(1, 2, 3, 4)
	assert.Equal(t, float64(10), result)
}

func TestPow(t *testing.T) {
	client, err := zbus.NewRedisClient("tcp://localhost:6379")
	if err != nil {
		panic(err)
	}

	calculator := stubs.NewCalculatorStub(client)

	result := calculator.Pow(2, 2)
	assert.Equal(t, float64(4), result)
}

func TestDivide(t *testing.T) {
	client, err := zbus.NewRedisClient("tcp://localhost:6379")
	if err != nil {
		panic(err)
	}

	calculator := stubs.NewCalculatorStub(client)

	_, err = calculator.Divide(2, 0)
	assert.Error(t, err, "division by 0 should return an error")

	result, err := calculator.Divide(4, 2)
	assert.NoError(t, err)
	assert.Equal(t, float64(2), result)
}

func TestAvg(t *testing.T) {
	client, err := zbus.NewRedisClient("tcp://localhost:6379")
	if err != nil {
		panic(err)
	}

	calculator := stubs.NewCalculatorStub(client)

	result := calculator.Avg([]float64{10, 10, 10})
	assert.Equal(t, float64(10), result)
}

func TestCapitalize(t *testing.T) {
	client, err := zbus.NewRedisClient("tcp://localhost:6379")
	if err != nil {
		panic(err)
	}

	utils := stubs.NewUtilsStub(client)

	result := utils.Capitalize("hello world")
	assert.Equal(t, result, "HELLO WORLD")
}

func TestTickTock(t *testing.T) {
	client, err := zbus.NewRedisClient("tcp://localhost:6379")
	if err != nil {
		panic(err)
	}

	utils := stubs.NewUtilsStub(client)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	now := time.Now()
	c, err := utils.TikTok(ctx)
	assert.NoError(t, err)
	results := []time.Time{}

	for t := range c {
		results = append(results, t)
	}
	assert.Len(t, results, 5)
	for _, result := range results {
		assert.True(t, result.Sub(now) > 0)
	}
}
