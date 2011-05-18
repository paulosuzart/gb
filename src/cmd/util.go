package main

import (
	"os"
)

func Min(x, y int64) int64 {
	if x == -1 {
		return y
	}

	if x < y {
		return x
	}
	return y

}

func Max(x, y int64) int64 {
	if x > y {
		return x
	}
	return y

}

type StringWritter struct {
	s string
}

func (self *StringWritter) Write(p []byte) (n int, err os.Error) {
	self.s += string(p)
	return len(self.s), nil
}
