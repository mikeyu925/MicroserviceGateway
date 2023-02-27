package loadbalance

import (
	"fmt"
	"testing"
)

func TestRoundRobin(t *testing.T) {
	rb := &RoundRobinBalance{}
	rb.Add("127.0.0.1:8000")
	rb.Add("127.0.0.1:8001")
	rb.Add("127.0.0.1:8002")
	rb.Add("127.0.0.1:8003")
	rb.Add("127.0.0.1:8004")
	rb.Add("127.0.0.1:8005")

	for i := 0; i < 10; i++ {
		fmt.Println(rb.Next())
	}
}
