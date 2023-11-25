package router

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mailru/easyjson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ujwegh/shortener/internal/app/config"
	"github.com/ujwegh/shortener/internal/app/handlers"
	"github.com/ujwegh/shortener/internal/app/middlware"
	"github.com/ujwegh/shortener/internal/app/model"
	"github.com/ujwegh/shortener/internal/app/service"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type MockStorage struct {
	urlMap   map[string]model.ShortenedURL
	userURLs []model.ShortenedURL
}

func (fss *MockStorage) DeleteUserURLs(ctx context.Context, userURL *uuid.UUID, shortURLKeys []string) error {
	//TODO implement me
	panic("implement me")
}

func (fss *MockStorage) CreateUserURL(ctx context.Context, userURL *model.UserURL) error {
	var shortenedURL model.ShortenedURL
	for _, url := range fss.urlMap {
		if url.UUID == userURL.ShortenedURLUUID {
			shortenedURL = url
		}
	}
	fss.userURLs = append(fss.userURLs, shortenedURL)
	return nil
}

func (fss *MockStorage) ReadUserURLs(ctx context.Context, uid *uuid.UUID) ([]model.ShortenedURL, error) {
	return fss.userURLs, nil
}

func (fss *MockStorage) Ping(ctx context.Context) error {
	return nil
}

func (fss *MockStorage) WriteShortenedURL(ctx context.Context, shortenedURL *model.ShortenedURL) error {
	fss.urlMap[shortenedURL.ShortURL] = *shortenedURL
	return nil
}

func (fss *MockStorage) ReadShortenedURL(ctx context.Context, shortURL string) (*model.ShortenedURL, error) {
	shortenedURL := fss.urlMap[shortURL]
	return &shortenedURL, nil
}

func (fss *MockStorage) WriteBatchShortenedURLSlice(ctx context.Context, slice []model.ShortenedURL) error {
	for _, shortenedURL := range slice {
		fss.urlMap[shortenedURL.ShortURL] = shortenedURL
	}
	return nil
}

func TestRequestZipper(t *testing.T) {
	// Setup
	c := config.AppConfig{}
	s := &MockStorage{
		urlMap:   make(map[string]model.ShortenedURL),
		userURLs: make([]model.ShortenedURL, 0),
	}
	ss := service.NewShortenerService(s)
	sh := handlers.NewShortenerHandlers(c.ShortenedURLAddr, 5, ss, s)
	tsc := service.NewTokenService(c)
	am := middlware.NewAuthMiddleware(tsc)
	router := NewAppRouter(sh, am)
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
