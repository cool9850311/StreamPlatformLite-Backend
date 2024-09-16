package infrastructure

import (
	"Go-Service/src/main/application/interface/stream"
	"Go-Service/src/main/infrastructure/initializer"
	"Go-Service/src/main/infrastructure/livestream"
	"testing"
	"time"
)

func TestLivestreamService(t *testing.T) {
	initializer.InitLog()
	// Setup

	var service stream.ILivestreamService = livestream.NewLivestreamService(initializer.Log)

	// Start the service
	err := service.StartService()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}

	// Run the service loop in a goroutine
	go func() {
		err := service.RunLoop()
		if err != nil {
			t.Errorf("RunLoop error: %v", err)
		}
	}()

	// Wait for the service to start
	time.Sleep(time.Second)

	// Open a stream
	service.OpenStream("test1", "1111111", "test", "test2")
	// Wait for 30 seconds
	time.Sleep(30 * time.Second)

	// Update the stream API key
	err = service.UpdateStreamOutPutPathUUID("1111111", "newApiKey")
	if err != nil {
		t.Fatalf("Failed to update stream API key: %v", err)
	}
	if err != nil {
		t.Fatalf("Failed to open stream: %v", err)
	}
	// Create an infinite loop
	for {
		time.Sleep(100 * time.Millisecond)
	}
}
