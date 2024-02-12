package main

import (
	"backend/orchestrator"
	"backend/orchestrator/db"
	"backend/orchestrator/endpoints/agent"
	"backend/orchestrator/endpoints/client"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	db.OpenDB()

	go startAgentServer()
	go orchestrator.RemoveInactive()

	startClientServer()
}

func startClientServer() {
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.DefaultCORSConfig))
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.POST("/add_expression", client.AddExpression)
	e.GET("/subscribe", client.Subscribe)
	e.POST("/apply_execution_time", client.ApplyExecutionTime)

	e.Logger.Fatal(e.Start(":1323"))
}

func startAgentServer() {
	e := echo.New()
	e.POST("/get_expression", agent.GetExpression)
	e.POST("/i_am_alive", agent.IAmAlive)
	e.POST("/submit_result", agent.SubmitResult)
	e.Logger.Fatal(e.Start(":1324"))
}
