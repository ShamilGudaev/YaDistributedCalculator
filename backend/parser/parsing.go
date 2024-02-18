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
	Left  SignedValue `parser:"@@"`
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
	Operator Operator    `parser:"@('*' | '/')"`
	Right    SignedValue `parser:"@@"`
}

type SignedValue struct {
	Sign  *Operator `parser:"(@('+' | '-'))?"`
	Value Value     `parser:"@@"`
}

func (s *SignedValue) AddToQueue(queue *[]Operation, executionTime map[Operator]time.Duration) Result {
	if s.Sign == nil {
		v := OpAdd
		s.Sign = &v
	}

	return s.Value.AddToQueue(queue, executionTime, *s.Sign)
}

type Value struct {
	Number        *float64 `parser:"  @(Float|Int)"`
	Subexpression *SumSub  `parser:"| '(' @@ ')'"`
}

type Zero struct{}

func (n *Zero) GetValue() float64 {
	return 0
}

var zero Zero

func (s *Value) AddToQueue(queue *[]Operation, executionTime map[Operator]time.Duration, sign Operator) Result {

	if s.Number != nil {
		v := *s.Number
		if sign == OpSub {
			v *= -1
		}

		return &Number{
			value: v,
		}
	}

	right := s.Subexpression.AddToQueue(queue, executionTime)

	o := &ArithmeticOp{
		value:         0,
		operator:      OpSub,
		left:          &zero,
		right:         right,
		executionTime: executionTime[OpSub],
	}

	*queue = append(*queue, o)
	return o
}

var Parser = participle.MustBuild[SumSub]()
