package main

import (
	"log"
)

func main() {
	portsService := NewPortsService()
	log.Println("Refreshing ports data...")
	if err := portsService.Refresh(false); err != nil {
		log.Fatalf("Failed to refresh ports data: %v", err)
	}

	for i := 0; i < 100; i++ {
		port, err := portsService.GetRandomUnassignedPort()
		if err != nil {
			log.Fatalf("Failed to get random unassigned port: %v", err)
		}
		log.Printf("Random unassigned port: %d", port)
	}
}
