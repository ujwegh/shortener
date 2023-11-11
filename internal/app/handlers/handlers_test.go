package handlers

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/mailru/easyjson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ujwegh/shortener/internal/app/model"
	"github.com/ujwegh/shortener/internal/app/service"
	"github.com/ujwegh/shortener/internal/app/storage"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type MockStorage struct {
	urlMap map[string]model.ShortenedURL
}

func (fss *MockStorage) Ping(ctx context.Context) error {
	return nil
}

func (fss *MockStorage) WriteShortenedURL(shortenedURL *model.ShortenedURL) error {
	fss.urlMap[shortenedURL.ShortURL] = *shortenedURL
	return nil
}

func (fss *MockStorage) ReadShortenedURL(shortURL string) (*model.ShortenedURL, error) {
	shortenedURL := fss.urlMap[shortURL]
	return &shortenedURL, nil
}

func TestUrlShortener_ShortenUrl(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name             string
		method           string
		route            string
		body             string
		shortenedURLAddr string
		want             want
	}{
		{
			name:             "positive shorten url test",
			route:            "/",
			method:           http.MethodPost,
			body:             "https://google.com",
			shortenedURLAddr: "http://localhost:8080",
			want: want{
				code:        201,
				response:    "http://localhost:8080/",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:             "empty route body",
			method:           http.MethodPost,
			route:            "/",
			body:             "",
			shortenedURLAddr: "http://localhost:8080",
			want: want{
				code:        400,
				contentType: "text/plain; charset=utf-8",
				response:    "Url is empty\n",
			},
		},
		{
			name:             "positive shorten url test",
			route:            "/",
			method:           http.MethodPost,
			body:             "https://google.com",
			shortenedURLAddr: "http://localhost:8090",
			want: want{
				code:        201,
				response:    "http://localhost:8090/",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			request := httptest.NewRequest(test.method, test.route, strings.NewReader(test.body))
			rctx := chi.NewRouteContext()
			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

			urlMap := make(map[string]model.ShortenedURL)
			storage := &MockStorage{urlMap: urlMap}
			us := &ShortenerHandlers{
				shortenerService: service.NewShortenerService(storage),
				shortenedURLAddr: test.shortenedURLAddr,
				storage:          storage,
			}
			us.ShortenURL(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			err2 := res.Body.Close()
			require.NoError(t, err2)

			if res.StatusCode == http.StatusCreated {
				split := strings.Split(string(body), test.want.response)
				assert.True(t, strings.Contains(test.want.response, split[0]))
				assert.True(t, len(split[1]) == 8)
				assert.Equal(t, 1, len(urlMap))
			} else {
				assert.Equal(t, test.want.response, string(body))
			}
		})
	}
}

func TestUrlShortener_APIShortenUrl(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name             string
		method           string
		route            string
		body             string
		shortenedURLAddr string
		want             want
	}{
		{
			name:             "positive shorten url test",
			route:            "/api/shorten",
			method:           http.MethodPost,
			body:             "{\"url\": \"https://google.com\"}",
			shortenedURLAddr: "http://localhost:8080",
			want: want{
				code:        201,
				response:    "http://localhost:8080/",
				contentType: "application/json",
			},
		},
		{
			name:             "empty body",
			method:           http.MethodPost,
			route:            "/api/shorten",
			body:             "",
			shortenedURLAddr: "http://localhost:8080",
			want: want{
				code:        400,
				contentType: "text/plain; charset=utf-8",
				response:    "Unable to parse body\n",
			},
		},
		{
			name:             "empty route body",
			method:           http.MethodPost,
			route:            "/api/shorten",
			body:             "{\"url\": \"\"}",
			shortenedURLAddr: "http://localhost:8080",
			want: want{
				code:        400,
				contentType: "text/plain; charset=utf-8",
				response:    "URL is empty\n",
			},
		},
		{
			name:             "positive shorten url test",
			route:            "/api/shorten",
			method:           http.MethodPost,
			body:             "{\"url\": \"https://google.com\"}",
			shortenedURLAddr: "http://localhost:8090",
			want: want{
				code:        201,
				response:    "http://localhost:8090/",
				contentType: "application/json",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			request := httptest.NewRequest(test.method, test.route, strings.NewReader(test.body))
			rctx := chi.NewRouteContext()
			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))
			request.Header.Set("Content-Type", "application/json")
			var urlMap = make(map[string]model.ShortenedURL)
			us := &ShortenerHandlers{
				shortenerService: service.NewShortenerService(&MockStorage{urlMap: urlMap}),
				shortenedURLAddr: test.shortenedURLAddr,
			}
			us.APIShortenURL(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			err2 := res.Body.Close()
			require.NoError(t, err2)

			if res.StatusCode == http.StatusCreated {
				var response = &ShortenResponseDto{}
				err := easyjson.Unmarshal(body, response)
				assert.Nil(t, err)
				split := strings.Split(response.Result, test.want.response)
				assert.True(t, strings.Contains(test.want.response, split[0]))
				assert.True(t, len(split[1]) == 8)
				assert.Equal(t, 1, len(urlMap))
			} else {
				assert.Equal(t, test.want.response, string(body))
			}
		})
	}
}

func TestURLShortener_HandleShortenedURL(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	var targetURL = "https://google.com"
	key := "hdINdIoD"
	wrongKey := "wrongKey"
	tests := []struct {
		pathVar string
		urlMap  map[string]model.ShortenedURL
		name    string
		method  string
		route   string
		body    string
		want    want
	}{
		{
			urlMap: map[string]model.ShortenedURL{
				key: {
					OriginalURL: targetURL,
				},
			},
			name:    "positive shorten url test",
			pathVar: key,
			route:   "/" + key,
			method:  http.MethodGet,
			want: want{
				code:        307,
				contentType: "text/html; charset=utf-8",
				response:    targetURL,
			},
		},
		{
			urlMap: map[string]model.ShortenedURL{
				key: {
					OriginalURL: targetURL,
				},
			},
			name:    "sent wrong key",
			pathVar: wrongKey,
			route:   "/" + wrongKey,
			method:  http.MethodGet,
			want: want{
				code:        404,
				contentType: "text/plain; charset=utf-8",
				response:    "Shortened url not found\n",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			request := httptest.NewRequest(test.method, test.route, strings.NewReader(test.body))
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", test.pathVar)
			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))
			us := &ShortenerHandlers{
				shortenerService: service.NewShortenerService(&MockStorage{urlMap: test.urlMap}),
			}
			us.HandleShortenedURL(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			err2 := res.Body.Close()
			require.NoError(t, err2)

			if res.StatusCode == http.StatusTemporaryRedirect {
				assert.Equal(t, test.want.response, res.Header.Get("Location"))
			} else {
				assert.Equal(t, test.want.response, string(body))
			}
		})
	}
}

func TestShortenerHandlers_Ping(t *testing.T) {
	type fields struct {
		shortenerService service.ShortenerService
		shortenedURLAddr string
		storage          storage.Storage
	}
	type args struct {
		writer  http.ResponseWriter
		request *http.Request
	}
	storage := &MockStorage{}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			name: "positive ping test",
			fields: fields{
				shortenerService: service.NewShortenerService(storage),
				shortenedURLAddr: "http://localhost:8080",
				storage:          storage,
			},
			args: args{
				writer:  httptest.NewRecorder(),
				request: httptest.NewRequest(http.MethodGet, "/ping", nil),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			us := &ShortenerHandlers{
				shortenerService: tt.fields.shortenerService,
				shortenedURLAddr: tt.fields.shortenedURLAddr,
				storage:          tt.fields.storage,
			}
			us.Ping(tt.args.writer, tt.args.request)
			// assert response
			res := tt.args.writer.(*httptest.ResponseRecorder).Result()
			res.Body.Close()
			assert.Equal(t, http.StatusOK, res.StatusCode)
		})
	}
}
