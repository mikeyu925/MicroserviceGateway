package main

import "fmt"

func maximalRectangle(matrix []string) int {
	if matrix == nil || len(matrix) == 0 {
		return 0
	}
	n, m := len(matrix), len(matrix[0])
	heights := make([]int, m)
	ans := 0
	for i := 0; i < n; i++ {
		for j, c := range matrix[i] {
			if c == '1' {
				heights[j] = heights[j] + 1
			} else {
				heights[j] = 0
			}
		}
		ans = max(ans, getMaxArea(heights))
	}
	return ans
}

func getMaxArea(heights []int) int {
	n := len(heights)
	st := []int{}
	ans := 0
	for i, h := range heights {
		for len(st) > 0 && h < heights[st[len(st)-1]] {
			v := heights[st[len(st)-1]]
			st = st[:len(st)-1]
			left := -1
			if len(st) > 0 {
				left = st[len(st)-1]
			}
			ans = max(ans, v*(i-left-1))
		}
		st = append(st, i)
	}
	for len(st) > 0 {
		v := heights[st[len(st)-1]]
		st = st[:len(st)-1]
		left := -1
		if len(st) > 0 {
			left = st[len(st)-1]
		}
		ans = max(ans, v*(n-left-1))
	}
	return ans
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func main() {
	ans := maximalRectangle([]string{"10100", "10111", "11111", "10010"})
	fmt.Println(ans)
}
