package oci

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type Manifest struct {
	SchemaVersion int          `json:"schemaVersion"`
	MediaType     string       `json:"mediaType"`
	Config        Descriptor   `json:"config"`
	Layers        []Descriptor `json:"layers"`

	Manifests []struct {
		Digest   string `json:"digest"`
		Platform struct {
			Architecture string `json:"architecture"`
			OS           string `json:"os"`
		} `json:"platform"`
	} `json:"manifests"`
}
type Descriptor struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
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

func PullManifestByDigest(imageName string, digest string) (*Manifest, error) {

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
		return nil, err
	}

	manifestURL := fmt.Sprintf("https://registry-1.docker.io/v2/library/%s/manifests/%s", imageName, digest)

	req, err := http.NewRequest("GET", manifestURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+tokenResp.Token)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	client := &http.Client{}
	manifestResp, err := client.Do(req)
	if err != nil {
		return nil, err
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

func DownloadLayer(imageName string, digest string) error {
	authURL := fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:library/%s:pull", imageName)

	resp, err := http.Get(authURL)
	if err != nil {
		return fmt.Errorf("ошибка получения токена: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	blobURL := fmt.Sprintf("https://registry-1.docker.io/v2/library/%s/blobs/%s", imageName, digest)
	req, err := http.NewRequest("GET", blobURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+tokenResp.Token)

	client := &http.Client{}
	blobResp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка скачивания слоя: %w", err)
	}
	defer blobResp.Body.Close()

	if blobResp.StatusCode != http.StatusOK {
		return fmt.Errorf("сервер вернул статус %d при скачивании слоя", blobResp.StatusCode)
	}

	cacheDir := "layers_cache"
	if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
		return fmt.Errorf("ошибка создания папки: %w", err)
	}

	safeFileName := fmt.Sprintf("%s.tar", digest[7:])
	filePath := filepath.Join(cacheDir, safeFileName)

	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("ошибка создания файла: %w", err)
	}
	defer out.Close()

	hasher := sha256.New()

	multiWriter := io.MultiWriter(out, hasher)

	fmt.Println("⏳ Начинаем загрузку слоя и подсчет хэша...")
	_, err = io.Copy(multiWriter, blobResp.Body)
	if err != nil {
		return fmt.Errorf("ошибка при сохранении файла: %w", err)
	}

	calculatedHash := "sha256:" + hex.EncodeToString(hasher.Sum(nil))

	if calculatedHash != digest {
		os.Remove(filePath)
		return fmt.Errorf("ошибка целостности! Ожидался %s, получили %s", digest, calculatedHash)
	}

	fmt.Printf("✅ Слой успешно сохранен: %s\n", filePath)
	return nil
}
