package router

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/mailru/easyjson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ujwegh/shortener/internal/app/config"
	"github.com/ujwegh/shortener/internal/app/handlers"
	"github.com/ujwegh/shortener/internal/app/service"
	"github.com/ujwegh/shortener/internal/app/storage"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestZipper(t *testing.T) {
	// Setup
	c := config.AppConfig{}
	s := storage.NewFileStorage(c.FileStoragePath)
	ss := service.NewShortenerService(s)
	sh := handlers.NewShortenerHandlers(c.ShortenedURLAddr, 2, ss, s)
	router := NewAppRouter(sh)
	ts := httptest.NewServer(router)
	defer ts.Close()

	type want struct {
		code            int
		response        string
		expectedHeaders map[string]string
	}
	tests := []struct {
		name             string
		method           string
		route            string
		body             map[string]string
		shortenedURLAddr string
		headers          map[string]string
		want             want
	}{
		{
			name:   "api shorten url test - gzip",
			route:  "/api/shorten",
			method: http.MethodPost,
			body: map[string]string{
				"url": "https://google.com",
			},
			shortenedURLAddr: "http://localhost:8090",
			headers: map[string]string{
				"Accept-Encoding": "gzip",
				"Content-Type":    "application/json",
			},
			want: want{
				code: 201,
				expectedHeaders: map[string]string{
					"Content-Type":     "application/json",
					"Content-Encoding": "gzip",
				},
			},
		},
		{
			name:   "api shorten url test - no gzip",
			route:  "/api/shorten",
			method: http.MethodPost,
			body: map[string]string{
				"url": "https://google.com",
			},
			shortenedURLAddr: "http://localhost:8090",
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			want: want{
				code: 201,
				expectedHeaders: map[string]string{
					"Content-Type": "application/json",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create a request to pass to our handler.
			payloadBytes, _ := json.Marshal(test.body)
			request, err := http.NewRequest(test.method, ts.URL+test.route, bytes.NewBuffer(payloadBytes))
			if err != nil {
				t.Fatalf("Could not create request: %v", err)
			}
			rctx := chi.NewRouteContext()
			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

			for headerName, headerValue := range test.headers {
				request.Header.Set(headerName, headerValue)
			}

			// Perform the HTTP request
			client := &http.Client{}
			resp, err := client.Do(request)
			if err != nil {
				t.Fatalf("Could not make request: %v", err)
			}
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					t.Fatalf("Could not close response body: %v", err)
				}
			}(resp.Body)

			// Check the response
			assert.Equal(t, test.want.code, resp.StatusCode)
			for headerKey, headerValue := range test.want.expectedHeaders {
				assert.Equal(t, headerValue, resp.Header.Get(headerKey), "header %s is not present", headerKey)
			}

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			err2 := resp.Body.Close()
			require.NoError(t, err2)

			if resp.StatusCode == http.StatusCreated {
				contendEncoding := resp.Header.Get("Content-Encoding")
				if strings.Contains(contendEncoding, "gzip") {
					var respDto = &handlers.ShortenResponseDto{}
					gr, err := gzip.NewReader(bytes.NewReader(body))
					require.NoError(t, err)
					err = easyjson.UnmarshalFromReader(gr, respDto)
					require.NoError(t, err)
				} else {
					var respDto = &handlers.ShortenResponseDto{}
					err = easyjson.Unmarshal(body, respDto)
					require.NoError(t, err)
				}
			} else {
				assert.Equal(t, test.want.response, string(body))
			}
		})
	}
}
