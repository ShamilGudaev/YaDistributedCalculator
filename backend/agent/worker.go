package agent

import (
	"backend/parser"
	pb "backend/proto"
	"time"
)

type Worker struct {
	ExpressionID uint64
	StopChannel  chan bool
}

type Result struct {
	ExpressionID uint64
	Result       float64
}

func (w Worker) Start(data *pb.GetExpressionResponse_GetExpressionResponseData, out chan Result) {
	p, err := parser.Parser.ParseString("", data.Expression)

	if err != nil {
		panic(err)
	}

	queue, result := p.CreateExecutionQueue(map[parser.Operator]time.Duration{
		parser.OpMul: time.Duration(data.OpMulMS) * time.Millisecond,
		parser.OpDiv: time.Duration(data.OpDivMS) * time.Millisecond,
		parser.OpAdd: time.Duration(data.OpAddMS) * time.Millisecond,
		parser.OpSub: time.Duration(data.OpSubMS) * time.Millisecond,
	})

	for _, o := range queue {
		executed := make(chan bool, 1)
		stopExecution := make(chan bool, 1)
		go o.Execute(executed, stopExecution)

		select {
		case <-w.StopChannel:
			stopExecution <- true
			return
		case <-executed:
			continue
		}
	}

	out <- Result{
		ExpressionID: data.ExpressionID,
		Result:       result.GetValue(),
	}
}
