package elasticsearch

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"

	elastic "github.com/olivere/elastic/v7"
)

var client *elastic.Client
var aiClient *elastic.Client

// GetESClient open connection
func GetESClient() (*elastic.Client, error) {
	var err error
	if client == nil {
		client, err = elastic.NewClient(
			elastic.SetURL(os.Getenv("ELASTICSEARCH_HOST")),
			elastic.SetSniff(false),
			elastic.SetHealthcheck(false))
	}
	return client, err
}

// GetAIESClient open connection AI
func GetAIESClient() (*elastic.Client, error) {
	var err error
	if aiClient == nil {
		aiClient, err = elastic.NewClient(
			elastic.SetURL(os.Getenv("AI_ELASTICSEARCH_HOST")),
			elastic.SetSniff(false),
			elastic.SetHealthcheck(false))
	}
	return aiClient, err
}

// DummyHTTPClient for mocking es responses
type DummyHTTPClient struct {
	handlers map[string]func(*http.Request) []byte
}

// HandleFunc adds handler for a specific url
func (c *DummyHTTPClient) HandleFunc(url string, handler func(*http.Request) []byte) {
	if c.handlers == nil {
		c.handlers = map[string]func(*http.Request) []byte{url: handler}
		return
	}
	c.handlers[url] = handler
}

// Do for handling request
func (c *DummyHTTPClient) Do(r *http.Request) (*http.Response, error) {

	handler, ok := c.handlers[r.URL.Path]
	if !ok {
		return nil, fmt.Errorf("invalid url")
	}

	response := handler(r)

	recorder := httptest.NewRecorder()
	recorder.Write(response)
	recorder.Header().Set("Content-Type", "application/json")

	return recorder.Result(), nil
}

// NewMockHTTPClient for handling responseMock
func NewMockHTTPClient() *DummyHTTPClient {
	return &DummyHTTPClient{}
}

// DummyElasticSearchClient returning *elastic.Client, error
func DummyElasticSearchClient(httpClient *DummyHTTPClient) (*elastic.Client, error) {
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	fmt.Printf("Elastic dummy test server now listening on %s ...", ts.URL)

	client, err := elastic.NewSimpleClient(
		elastic.SetURL(ts.URL),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false),
		elastic.SetHttpClient(httpClient))

	return client, err
}
