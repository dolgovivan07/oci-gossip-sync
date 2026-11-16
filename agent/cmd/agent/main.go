package main

import (
	"fmt"
	"log"

	"github.com/dolgovivan07/oci-gossip-sync/agent/pkg/oci"
)

func main() {
	fmt.Println("🚀 Запуск OCI Gossip Agent...")
	fmt.Println("1. Запрашиваем оглавление для образа alpine:latest...")

	index, err := oci.PullManifest("alpine", "latest")
	if err != nil {
		log.Fatalf("Ошибка: %v", err)
	}

	var targetDigest string

	for _, m := range index.Manifests {
		if m.Platform.OS == "linux" && m.Platform.Architecture == "amd64" {
			targetDigest = m.Digest
			break
		}
	}

	if targetDigest == "" {
		log.Fatalf("Не найдена архитектура linux/amd64 для этого образа")
	}

	fmt.Printf("✅ Найдена архитектура linux/amd64. Хэш: %s\n", targetDigest)
	fmt.Println("2. Запрашиваем слои для этой архитектуры...")

	manifest, err := oci.PullManifestByDigest("alpine", targetDigest)
	if err != nil {
		log.Fatalf("Ошибка получения манифеста слоев: %v", err)
	}

	fmt.Printf("✅ Успех! Найдено слоев: %d\n", len(manifest.Layers))

	for i, layer := range manifest.Layers {
		fmt.Printf("Слой %d:\n", i+1)
		fmt.Printf("  Размер: %d байт\n", layer.Size)
		fmt.Printf("  Хэш:    %s\n", layer.Digest)
	}
	if len(manifest.Layers) > 0 {
		layerDigest := manifest.Layers[0].Digest
		fmt.Println("\n3. Скачиваем первый слой...")

		err = oci.DownloadLayer("alpine", layerDigest)
		if err != nil {
			log.Fatalf("Ошибка при скачивании слоя: %v", err)
		}

		fmt.Println("🚀 ЭТАП 1 УСПЕШНО ЗАВЕРШЕН!")
	}
}
