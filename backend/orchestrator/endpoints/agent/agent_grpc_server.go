package agent

import (
	pb "backend/proto"
)

type AgentGRPCServer struct {
	pb.UnimplementedOrchestratorServer
}
