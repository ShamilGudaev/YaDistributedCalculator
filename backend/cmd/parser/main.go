package main

import (
	"backend/parser"
	"fmt"
	"time"
)

func main() {
	src := "1 + -2 + 4 * -(8 + 16)"
	println(src)
	res, err := parser.Parser.ParseString("", src)
	if err != nil {
		panic(err)
	}

	queue := make([]parser.Operation, 0)
	executionTime := make(map[parser.Operator]time.Duration)
	executionTime[parser.OpMul] = 0
	executionTime[parser.OpDiv] = 0
	executionTime[parser.OpAdd] = 0
	executionTime[parser.OpSub] = 0
	r := res.AddToQueue(&queue, executionTime)

	c := make(chan bool, 1)
	esc := make(chan bool, 1)

	for _, o := range queue {
		o.Execute(c, esc)
		<-c
	}

	fmt.Printf("Result: %f\n", r.GetValue())
}
