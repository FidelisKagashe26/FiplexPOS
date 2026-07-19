package main

import (
	_ "POS-fiplex/docs"
	"POS-fiplex/server"
)

// @title POS Fiplex API
// @version 1.0
// @description POS Fiplex API
// @host localhost:8080
// @BasePath /api/v1
func main() {
	app := server.InitApp()
	defer server.Cleanup(app)

	go server.StartServer(app)

	server.WaitForShutdown(app)
}
