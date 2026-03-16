package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please provide an environment ID")
	}
	envID := os.Args[1]
	url := fmt.Sprintf("http://localhost:8084/stream?environment_id=%s", envID)
	
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error connecting to stream: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("Connected to stream for environment %s\n", envID)
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Error reading from stream: %v", err)
		}
		fmt.Print(line)
	}
}
