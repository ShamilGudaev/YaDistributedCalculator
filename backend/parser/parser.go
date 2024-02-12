package parser

import (
	"time"

	"github.com/alecthomas/participle/v2"
)

type Operator int

const (
	OpMul Operator = iota
	OpDiv
	OpAdd
	OpSub
)

var operatorMap = map[string]Operator{"+": OpAdd, "-": OpSub, "*": OpMul, "/": OpDiv}

func (o *Operator) Capture(s []string) error {
	*o = operatorMap[s[0]]
	return nil
}

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
		o.value = o.left.GetValue() + o.right.GetValue()
	case OpSub:
		o.value = o.left.GetValue() - o.right.GetValue()
	case OpMul:
		o.value = o.left.GetValue() * o.right.GetValue()
	case OpDiv:
		o.value = o.left.GetValue() / o.right.GetValue()
	}

	o.left = nil
	o.right = nil

	executed <- true
}

func (o *ArithmeticOp) GetValue() float64 {
	return o.value
}

type SumSub struct {
	Left  MulDiv      `parser:"@@"`
	Right []*OpSumSub `parser:"@@*"`
}

func (s *SumSub) CreateExecutionQueue(executionTime map[Operator]time.Duration) (queue []Operation, result Result) {
	queue = make([]Operation, 0)
	result = s.AddToQueue(&queue, executionTime)
	return
}

func (s *SumSub) AddToQueue(queue *[]Operation, executionTime map[Operator]time.Duration) Result {
	var left = s.Left.AddToQueue(queue, executionTime)

	for _, v := range s.Right {
		o := &ArithmeticOp{
			value:         0,
			operator:      v.Operator,
			left:          left,
			right:         v.Right.AddToQueue(queue, executionTime),
			executionTime: executionTime[v.Operator],
		}

		*queue = append(*queue, o)
		left = o
	}

	return left
}

type OpSumSub struct {
	Operator Operator `parser:"@('+' | '-')"`
	Right    MulDiv   `parser:"@@"`
}

type MulDiv struct {
	Left  Value       `parser:"@@"`
	Right []*OpMulDiv `parser:"@@*"`
}

func (s *MulDiv) AddToQueue(queue *[]Operation, executionTime map[Operator]time.Duration) Result {
	var left = s.Left.AddToQueue(queue, executionTime)

	for _, v := range s.Right {
		o := &ArithmeticOp{
			value:         0,
			operator:      v.Operator,
			left:          left,
			right:         v.Right.AddToQueue(queue, executionTime),
			executionTime: executionTime[v.Operator],
		}

		*queue = append(*queue, o)
		left = o
	}

	return left
}

type OpMulDiv struct {
	Operator Operator `parser:"@('*' | '/')"`
	Right    Value    `parser:"@@"`
}

type Value struct {
	Number        *float64 `parser:"  @(Float|Int)"`
	Subexpression *SumSub  `parser:"| '(' @@ ')'"`
}

func (s *Value) AddToQueue(queue *[]Operation, executionTime map[Operator]time.Duration) Result {
	if s.Number != nil {
		return &Number{
			value: *s.Number,
		}
	}

	return s.Subexpression.AddToQueue(queue, executionTime)
}

// func (expr *OpSumSub) MakeOperationSequence() ([]Computable, Result) {
// 	queue := make([]Computable, 0)
// 	result := expr.AddComputable(&queue)
// 	return queue, result
// }

var Parser = participle.MustBuild[SumSub]()
