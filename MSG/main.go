package main

import "fmt"

type node struct {
	key, val int
}

func main() {
	nodes := make([]node, 4)
	for i := 0; i < 4; i++ {
		nodes[i] = node{i, i * 2}
	}
	for i := 0; i < 4; i++ {
		fmt.Printf("%p,%v\n", &nodes[i], nodes[i])
	}

	fmt.Println("-------------")
	for _, v := range nodes {
		fmt.Printf("%p,%v\n", &v, v)
	}
}
