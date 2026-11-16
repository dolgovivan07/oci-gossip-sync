package oci

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Manifest struct {
	SchemaVersion int          `json:"schemaVersion"`
	MediaType     string       `json:"mediaType"`
	Config        Descriptor   `json:"config"`
	Layers        []Descriptor `json:"layers"`
}
type Descriptor struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"` // Тот самый хэш слоя (например, sha256:abcd...)
	Size      int64  `json:"size"`   // int64, потому что слои могут быть большими (гигабайты)
}

func PullManifest(imageName string, tag string) (*Manifest, error) {
	authURL := fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:library/%s:pull", imageName)
	resp, err := http.Get(authURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения токена: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("ошибка чтения токена: %w", err)

	}

	manifestURL := fmt.Sprintf("https://registry-1.docker.io/v2/library/%s/manifests/%s", imageName, tag)

	req, err := http.NewRequest("GET", manifestURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+tokenResp.Token)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	client := &http.Client{}
	manifestResp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка скачивания манифеста: %w", err)
	}
	defer manifestResp.Body.Close()

	if manifestResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(manifestResp.Body)
		return nil, fmt.Errorf("ошибка registry %d: %s", manifestResp.StatusCode, string(bodyBytes))
	}
	var manifest Manifest
	if err := json.NewDecoder(manifestResp.Body).Decode(&manifest); err != nil {
		return nil, err
	}
	return &manifest, nil

}
