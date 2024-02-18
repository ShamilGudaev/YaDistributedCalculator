package parser

import (
	"time"
)

type Result interface {
	GetValue() float64
}

type Number struct {
	value float64
}

func (n *Number) GetValue() float64 {
	return n.value
}

type Operation interface {
	Execute(executed chan bool, stopExecution chan bool)
}

type ArithmeticOp struct {
	operator      Operator
	value         float64
	left          Result
	right         Result
	executionTime time.Duration
}

func (o *ArithmeticOp) Execute(executed chan bool, stopExecution chan bool) {
	if o.left == nil || o.right == nil {
		panic("Already executed")
	}

	done := make(chan bool, 1)
	go func() {
		time.Sleep(o.executionTime)
		done <- true
	}()

	select {
	case <-stopExecution:
		return
	case <-done:
		break
	}

	switch o.operator {
	case OpAdd:
		// fmt.Printf("%f + %f == %f\n", o.left.GetValue(), o.right.GetValue(), o.left.GetValue()+o.right.GetValue())
		o.value = o.left.GetValue() + o.right.GetValue()
	case OpSub:
		// fmt.Printf("%f - %f == %f\n", o.left.GetValue(), o.right.GetValue(), o.left.GetValue()-o.right.GetValue())
		o.value = o.left.GetValue() - o.right.GetValue()
	case OpMul:
		// fmt.Printf("%f * %f == %f\n", o.left.GetValue(), o.right.GetValue(), o.left.GetValue()*o.right.GetValue())
		o.value = o.left.GetValue() * o.right.GetValue()
	case OpDiv:
		// fmt.Printf("%f / %f == %f\n", o.left.GetValue(), o.right.GetValue(), o.left.GetValue()/o.right.GetValue())
		o.value = o.left.GetValue() / o.right.GetValue()
	}

	o.left = nil
	o.right = nil

	executed <- true
}

func (o *ArithmeticOp) GetValue() float64 {
	return o.value
}
