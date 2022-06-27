package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/threefoldtech/zbus"
	"github.com/threefoldtech/zbus/examples/server/api"
)

type Calculator struct{}
type Utils struct{}

var (
	_ api.Calculator = (*Calculator)(nil)
	_ api.Utils      = (*Utils)(nil)
)

func (c *Calculator) Add(a, b float64) float64 {
	return a + b
}

func (c *Calculator) AddSub(a, b float64) (float64, float64) {
	return a + b, a - b
}

func (c *Calculator) Pow(a, b float64) float64 {
	return math.Pow(a, b)
}

func (c *Calculator) Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("division by zero")
	}

	return a / b, nil
}

func (c *Calculator) Avg(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	var t float64
	for _, x := range v {
		t += x
	}
	return t / float64(len(v))
}

func (u *Utils) Capitalize(s string) string {
	return strings.ToUpper(s)
}

func (u *Utils) Tuple() (int, string, float64, error) {
	return 10, "hello world", 0.5, nil
}

func (u *Utils) TikTok(ctx context.Context) <-chan time.Time {
	c := make(chan time.Time)

	go func() {
		ticker := time.NewTicker(time.Second)
		defer close(c)
		defer ticker.Stop()

		for instance := range ticker.C {
			select {
			case <-ctx.Done():
				return
			case c <- instance:
			}
		}
	}()

	return c
}

func (u *Utils) Sleep(t time.Duration) error {
	fmt.Printf("sleeping for %s\n", t)
	<-time.After(t)
	return nil
}

func (u *Utils) Panic() int {
	panic("Aaaaah!")

}

func main() {
	server, err := zbus.NewRedisServer("server", "tcp://localhost:6379", 1)

	if err != nil {
		panic(err)
	}
	var calc Calculator
	var utils Utils

	server.Register(zbus.ObjectID{Name: "calculator", Version: "1.0"}, &calc)
	server.Register(zbus.ObjectID{Name: "utils", Version: "1.0"}, &utils)

	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)
	go func() {
		<-ch
		log.Println("shutting down!")
		cancel()
	}()

	if err := server.Run(ctx); err != nil && err != context.Canceled {
		panic(err)
	}
}
