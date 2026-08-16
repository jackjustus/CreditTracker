package embed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// Ollama's default listen address, used only when a binary is run directly on
// the host (`go run ./cmd/embedder`) against a locally installed daemon. Under
// compose, Ollama is a container and OLLAMA_URL points at the service.
const defaultURL = "http://localhost:11434"

const (
	Model      = "nomic-embed-text"
	Dimensions = 768
)

type Client struct {
	url   string
	httpc *http.Client
}

func NewClient() (*Client, error) {
	url := os.Getenv("OLLAMA_URL")
	if url == "" {
		url = defaultURL
	}
	return &Client{
		url:   url,
		httpc: &http.Client{Timeout: 5 * time.Second},
	}, nil
}

type embedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embedResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
	Error      string      `json:"error"`
}

// Embed takes in a list of strings, and returns vectors for each input string
func (c *Client) Embed(ctx context.Context, docs []string) ([][]float32, error) {
	if len(docs) == 0 {
		return nil, nil
	}

	body, err := json.Marshal(embedRequest{Model: Model, Input: docs})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url+"/api/embed", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	res, err := c.httpc.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal(err)
			return
		}
	}(res.Body)

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var parsed embedResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("ollama returned %s: %s", res.Status, raw)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned %s: %s", res.Status, parsed.Error)
	}

	if len(parsed.Embeddings) != len(docs) {
		return nil, fmt.Errorf("asked for %d embeddings, got %d", len(docs), len(parsed.Embeddings))
	}
	return parsed.Embeddings, nil
}

func (c *Client) EmbedOne(ctx context.Context, doc string) ([]float32, error) {
	vecs, err := c.Embed(ctx, []string{doc})
	if err != nil {
		return nil, err
	}
	return vecs[0], nil
}
