package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/JustWorking42/shortener-go-yandex/internal/app/models"
	"github.com/go-resty/resty/v2"
)

func main() {
	client := resty.New()

	resp, err := client.R().
		SetHeader("Accept-Encoding", "gzip").
		Get("http://localhost:8080/ping")
	if err != nil {
		log.Fatalf("Error sending GET /ping request: %v", err)
	}
	fmt.Println("GET /ping Response: ", resp)

	body := []byte("https://calendar.google.com")
	compressedBody, err := compressData(body)
	if err != nil {
		log.Fatalf("Error compressing data: %v", err)
	}

	postResp, err := client.R().
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("Content-Encoding", "gzip").
		SetBody(compressedBody).
		Post("http://localhost:8080/")
	if err != nil {
		log.Fatalf("Error sending POST / request: %v", err)
	}
	fmt.Println("POST / Response: ", postResp)

	id := strings.TrimPrefix(postResp.String(), "http://localhost:8080/")
	fmt.Println("POST / Id: ", id)

	getResp, err := client.R().
		SetHeader("Accept-Encoding", "gzip").
		Get(fmt.Sprintf("http://localhost:8080/%s", id))
	if err != nil {
		log.Fatalf("Error sending GET /%s request: %v", id, err)
	}
	fmt.Println(fmt.Sprintf("GET /%s Response: ", id), getResp)

	req := &models.RequestShotenerURL{
		URL: "https://example.com",
	}

	jsonReq, err := json.Marshal(req)
	if err != nil {
		log.Fatalf("Error marshalling RequestShotenerURL to JSON: %v", err)
	}

	body = []byte(jsonReq)
	compressedBody, err = compressData(body)
	if err != nil {
		log.Fatalf("Error compressing data: %v", err)
	}

	shortenResp, err := client.R().
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("Content-Encoding", "gzip").
		SetBody(compressedBody).
		Post("http://localhost:8080/api/shorten")
	if err != nil {
		log.Fatalf("Error sending POST /api/shorten request: %v", err)
	}
	fmt.Println("POST /api/shorten Response: ", shortenResp)

	reqs := []models.RequestShortenerURLBatch{
		{
			ID:  "id1",
			URL: "https://example1.com",
		},
		{
			ID:  "id2",
			URL: "https://example2.com",
		},
	}

	jsonReqs, err := json.Marshal(reqs)
	if err != nil {
		log.Fatalf("Error marshalling RequestShortenerURLBatch to JSON: %v", err)
	}

	body = []byte(jsonReqs)
	compressedBody, err = compressData(body)
	if err != nil {
		log.Fatalf("Error compressing data: %v", err)
	}

	batchResp, err := client.R().
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("Content-Encoding", "gzip").
		SetBody(compressedBody).
		Post("http://localhost:8080/api/shorten/batch")
	if err != nil {
		log.Fatalf("Error sending POST /api/shorten/batch request: %v", err)
	}
	fmt.Println("POST /api/shorten/batch Response: ", batchResp)

	userUrlsResp, err := client.R().
		SetHeader("Accept-Encoding", "gzip").
		Get("http://localhost:8080/api/user/urls")
	if err != nil {
		log.Fatalf("Error sending GET /api/user/urls request: %v", err)
	}
	fmt.Println("GET /api/user/urls Response: ", userUrlsResp)

	deleteResp, err := client.R().
		SetHeader("Accept-Encoding", "gzip").
		Delete(fmt.Sprintf("http://localhost:8080/api/user/urls/%s", id))
	if err != nil {
		log.Fatalf("Error sending DELETE /api/user/urls/%s request: %v", id, err)
	}
	fmt.Println(fmt.Sprintf("DELETE /api/user/urls/%s Response: ", id), deleteResp)
}
func compressData(data []byte) ([]byte, error) {
	// Create a buffer to hold the compressed data
	var buf bytes.Buffer

	// Create a gzip writer
	gz := gzip.NewWriter(&buf)

	// Write the data to the gzip writer
	_, err := gz.Write(data)
	if err != nil {
		return nil, fmt.Errorf("error writing data to gzip writer: %w", err)
	}

	// Close the gzip writer
	err = gz.Close()
	if err != nil {
		return nil, fmt.Errorf("error closing gzip writer: %w", err)
	}

	// Return the compressed data
	return buf.Bytes(), nil
}
