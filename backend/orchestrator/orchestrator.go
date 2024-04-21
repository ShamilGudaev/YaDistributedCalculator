package orchestrator

import (
	"backend/orchestrator/cfg"
	"backend/orchestrator/db"
	"backend/orchestrator/endpoints/agent"
	"backend/orchestrator/endpoints/client"
	"backend/orchestrator/middleware"
	"log"
	"net"

	pb "backend/proto"

	"github.com/labstack/echo/v4"
	mw "github.com/labstack/echo/v4/middleware"
	"google.golang.org/grpc"
)

func StartOrchestrator() {
	cfg.Init()
	db.OpenDB()

	go startAgentServer()
	go RemoveInactive()

	startClientServer()
}

func startClientServer() {
	e := echo.New()

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Before(func() {
				c.Response().Header().Set(echo.HeaderAccessControlAllowOrigin, "http://127.0.0.1:5173")
				c.Response().Header().Set(echo.HeaderAccessControlAllowCredentials, "true")
				c.Response().Header().Set(echo.HeaderAccessControlAllowMethods, "GET, POST")
			})
			return next(c)
		}
	})

	e.Use(mw.CORSWithConfig(mw.CORSConfig{
		AllowOrigins: []string{"http://127.0.0.1:5173"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		AllowMethods: []string{"POST", "GET"},
	}))

	e.Use(middleware.SetUserIDMiddleware)

	e.POST("/add_expression", client.AddExpression, middleware.CheckIfAuthorized)
	e.GET("/subscribe", client.Subscribe, middleware.CheckIfAuthorized)
	e.POST("/apply_execution_time", client.ApplyExecutionTime, middleware.CheckIfAuthorized)

	e.POST("/registration", client.UserRegistration, middleware.CheckIfNotAuthorized)
	e.POST("/authorization", client.UserAuthorization, middleware.CheckIfNotAuthorized)

	e.POST("/am_i_authorized", client.AmIAuthorized)
	e.POST("/logout", client.Logout)

	e.Logger.Fatal(e.Start(":1323"))
}

func startAgentServer() {
	// e := echo.New()
	// e.POST("/get_expression", agent.GetExpression)
	// e.POST("/i_am_alive", agent.IAmAlive)
	// e.POST("/submit_result", agent.SubmitResult)
	// e.Logger.Fatal(e.Start(":1324"))

	lis, err := net.Listen("tcp", "0.0.0.0:1324")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterOrchestratorServer(s, &agent.AgentGRPCServer{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
