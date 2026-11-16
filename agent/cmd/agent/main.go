package main

import (
	"fmt"
	"log"

	"github.com/dolgovivan07/oci-gossip-sync/agent/pkg/oci"
)

func main() {
	fmt.Println("🚀 Запуск OCI Gossip Agent...")
	fmt.Println("Запрашиваем манифест для образа alpine:latest...")

	manifest, err := oci.PullManifest("alpine", "latest")

	if err != nil {
		log.Fatalf("Критическая ошибка: %v", err)
	}

	fmt.Println("✅ Манифест успешно получен!")
	fmt.Printf("Количество слоев: %d\n", len(manifest.Layers))
	for i, layer := range manifest.Layers {
		fmt.Printf("Слой %d: %d байт | Хэш: %s\n", i+1, layer.Size, layer.Digest)
	}
}
