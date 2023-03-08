package loadbalance

import (
	"fmt"
	"math/rand"
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

func TestWeightRoundRobin(t *testing.T) {
	rb := &WeightRoundRobinBalance{}

	rb.Add("127.0.0.1:8000", "6")
	rb.Add("127.0.0.1:8001", "3")
	rb.Add("127.0.0.1:8002", "1")
	print(rb, "")
	fmt.Println("--------初始化完成-------")
	for i := 0; i < 20; i++ {
		addr, err := rb.Next()
		if err != nil {
			fmt.Println(err)
		}
		var r = rand.Intn(3)
		// 模拟故障
		if r == 1 {
			fmt.Println("server " + addr + " has failed!")
			rb.CallBack(addr, false)
		} else {
			fmt.Println("server " + addr + " success!")
			rb.CallBack(addr, true)
		}

		print(rb, addr)
	}
}

func TestRandomRoundRobin(t *testing.T) {
	rb := &RandomBalance{}
	rb.Add("127.0.0.1:8000")
	rb.Add("127.0.0.1:8001")
	rb.Add("127.0.0.1:8002")
	rb.Add("127.0.0.1:8003")
	rb.Add("127.0.0.1:8004")
	rb.Add("127.0.0.1:8005")
	fmt.Println("--------初始化完成-------")
	for i := 0; i < 20; i++ {
		fmt.Println(rb.Next())
	}
}

func TestConsistentHashRoundRobin(t *testing.T) {
	rb := NewConsistentHashBalance(2, nil)
	rb.Add("127.0.0.1:8000")
	rb.Add("127.0.0.1:8001")
	rb.Add("127.0.0.1:8002")
	rb.Add("127.0.0.1:8003")
	rb.Add("127.0.0.1:8004")
	rb.Add("127.0.0.1:8005")
	fmt.Println(rb.hashKeys)
	fmt.Println(rb.Get("http://127.0.0.1:8002/demo/get"))
	fmt.Println(rb.Get("http://127.0.0.1:8003/demo/get"))
	fmt.Println(rb.Get("http://127.0.0.1:8002/demo/get"))
	fmt.Println(rb.Get("http://127.0.0.1:8003/demo/put"))
	fmt.Println(rb.Get("http://127.0.0.1:8001/demo/get"))

	//
	fmt.Println("---------------")
	fmt.Println(rb.Get("127.0.0.1:8002"))
	fmt.Println(rb.Get("127.0.0.1:8001"))
	fmt.Println(rb.Get("127.0.0.1:8002"))
	fmt.Println(rb.Get("127.0.0.1:8003"))
	fmt.Println(rb.Get("127.0.0.1:8004"))
	fmt.Println(rb.Get("127.0.0.1:8005"))
	fmt.Println(rb.Get("127.0.0.1:8008"))
	fmt.Println(rb.Get("127.0.0.1:8004"))

}
