package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/swaggo/echo-swagger"
	"tpoints/handler"
	
	_ "tpoints/docs"
)

// @title TPoints
// @version 1.0
// @description 

// @contact.name Kiswono Prayogo

// @host 127.0.0.1:1323
// @BasePath /
func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	
	// init dependencies
	server := handler.InitServer()

	// documentation
	e.GET(`/swagger/*`, echoSwagger.WrapHandler)
	
	// APIs
	e.GET(`/points/logs/:userId/:limit/:offset`,server.Wrap(handler.Points_Logs))
	e.POST(`/points/queue/:userId/:delta`, server.Wrap(handler.Points_Queue))
	e.POST(`/points/add/:userId/:delta`,server.Wrap(handler.Points_Add))

	e.Logger.Fatal(e.Start(":1323"))
}
