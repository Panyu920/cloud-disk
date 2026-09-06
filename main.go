package main

import (
	"log"

	"github.com/Panyu920/cloud-disk/server"
)

func main() {
	s := server.New(":8080")

	// 启动服务
	log.Println("Starting server on :8080")
	if s.Start() != nil {
		log.Fatal("Failed to start server")
	}
}
