package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	// Позже мы добавим сюда парсинг флагов командной строки (например, порт или адрес registry)
	fmt.Println("🚀 Starting OCI Gossip Agent...")

	// Читаем порт из переменных окружения или ставим дефолтный 8080
	port := os.Getenv("AGENT_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Agent is ready and listening on port %s (stub)", port)
}
