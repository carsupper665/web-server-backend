package main

import (
	"log"

	"github.com/example/minecraft-server-controller/internal/router"
)

func main() {
	r := router.SetupRouter()
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
