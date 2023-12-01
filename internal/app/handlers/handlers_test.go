package handlers

import (
	"context"
	"database/sql"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mailru/easyjson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appContext "github.com/ujwegh/shortener/internal/app/context"
	"github.com/ujwegh/shortener/internal/app/model"
	"github.com/ujwegh/shortener/internal/app/service"
	"github.com/ujwegh/shortener/internal/app/storage"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type MockStorage struct {
	urlMap   map[string]model.ShortenedURL
	userURLs []model.ShortenedURL
}

func (fss *MockStorage) DeleteBulk(background context.Context, buffer map[uuid.UUID][]string) error {
	return nil
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

func TestUrlShortener_ShortenUrl(t *testing.T) {
	userUID := uuid.New()

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
		contextTimeout   time.Duration
		want             want
	}{
		{
			name:             "positive shorten url test",
			route:            "/",
			method:           http.MethodPost,
			body:             "https://google.com",
			shortenedURLAddr: "http://localhost:8080",
			contextTimeout:   time.Duration(2) * time.Second,
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
			contextTimeout:   time.Duration(2) * time.Second,
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
			contextTimeout:   time.Duration(2) * time.Second,
			want: want{
				code:        201,
				response:    "http://localhost:8090/",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:             "context timeout",
			route:            "/",
			method:           http.MethodPost,
			body:             "https://google.com",
			shortenedURLAddr: "http://localhost:8090",
			contextTimeout:   time.Duration(0) * time.Second,
			want: want{
				code:        500,
				contentType: "text/plain; charset=utf-8",
				response:    "Timeout exceeded\n",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			request := httptest.NewRequest(test.method, test.route, strings.NewReader(test.body))
			rctx := chi.NewRouteContext()
			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))
			request = request.WithContext(appContext.WithUserUID(request.Context(), &userUID))

			urlMap := make(map[string]model.ShortenedURL)
			storage := &MockStorage{urlMap: urlMap}
			us := &ShortenerHandlers{
				shortenerService: service.NewShortenerService(storage, nil),
				shortenedURLAddr: test.shortenedURLAddr,
				storage:          storage,
				contextTimeout:   test.contextTimeout,
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
	userUID := uuid.New()

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
		contextTimeout   time.Duration
		want             want
	}{
		{
			name:             "positive shorten url test",
			route:            "/api/shorten",
			method:           http.MethodPost,
			body:             "{\"url\": \"https://google.com\"}",
			shortenedURLAddr: "http://localhost:8080",
			contextTimeout:   time.Duration(2) * time.Second,
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
			contextTimeout:   time.Duration(2) * time.Second,
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
			contextTimeout:   time.Duration(2) * time.Second,
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
			contextTimeout:   time.Duration(2) * time.Second,
			want: want{
				code:        201,
				response:    "http://localhost:8090/",
				contentType: "application/json",
			},
		},
		{
			name:             "context timeout",
			route:            "/api/shorten",
			method:           http.MethodPost,
			body:             "{\"url\": \"https://google.com\"}",
			shortenedURLAddr: "http://localhost:8090",
			contextTimeout:   time.Duration(0) * time.Second,
			want: want{
				code:        500,
				contentType: "text/plain; charset=utf-8",
				response:    "Timeout exceeded\n",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			request := httptest.NewRequest(test.method, test.route, strings.NewReader(test.body))
			rctx := chi.NewRouteContext()
			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))
			request = request.WithContext(appContext.WithUserUID(request.Context(), &userUID))

			request.Header.Set("Content-Type", "application/json")
			var urlMap = make(map[string]model.ShortenedURL)
			s := &MockStorage{urlMap: urlMap}
			us := &ShortenerHandlers{
				shortenerService: service.NewShortenerService(s, make(chan service.Task)),
				shortenedURLAddr: test.shortenedURLAddr,
				storage:          s,
				contextTimeout:   test.contextTimeout,
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
	key2 := "dUBdl93n"
	tests := []struct {
		name           string
		pathVar        string
		urlMap         map[string]model.ShortenedURL
		method         string
		route          string
		body           string
		contextTimeout time.Duration
		want           want
	}{
		{
			name: "positive shorten url test",
			urlMap: map[string]model.ShortenedURL{
				key: {
					OriginalURL: targetURL,
				},
			},
			pathVar:        key,
			route:          "/" + key,
			method:         http.MethodGet,
			contextTimeout: time.Duration(2) * time.Second,
			want: want{
				code:        307,
				contentType: "text/html; charset=utf-8",
				response:    targetURL,
			},
		},
		{
			name: "sent wrong key",
			urlMap: map[string]model.ShortenedURL{
				key: {
					OriginalURL: targetURL,
				},
			},
			pathVar:        wrongKey,
			route:          "/" + wrongKey,
			method:         http.MethodGet,
			contextTimeout: time.Duration(2) * time.Second,
			want: want{
				code:        404,
				contentType: "text/plain; charset=utf-8",
				response:    "Shortened url not found\n",
			},
		},
		{
			name: "deleted url",
			urlMap: map[string]model.ShortenedURL{
				key2: {
					OriginalURL: targetURL,
					DeletedFlag: true,
				},
			},
			pathVar:        key2,
			route:          "/" + key2,
			method:         http.MethodGet,
			contextTimeout: time.Duration(2) * time.Second,
			want: want{
				code: 410,
			},
		},
		{
			name: "context timeout",
			urlMap: map[string]model.ShortenedURL{
				key: {
					OriginalURL: targetURL,
				},
			},
			pathVar:        key,
			route:          "/" + key,
			method:         http.MethodGet,
			contextTimeout: time.Duration(0) * time.Second,
			want: want{
				code:        500,
				contentType: "text/plain; charset=utf-8",
				response:    "Timeout exceeded\n",
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
				shortenerService: service.NewShortenerService(&MockStorage{urlMap: test.urlMap}, make(chan service.Task)),
				contextTimeout:   test.contextTimeout,
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
				shortenerService: service.NewShortenerService(storage, make(chan service.Task)),
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
				contextTimeout:   time.Duration(2) * time.Second,
			}
			us.Ping(tt.args.writer, tt.args.request)
			// assert response
			res := tt.args.writer.(*httptest.ResponseRecorder).Result()
			res.Body.Close()
			assert.Equal(t, http.StatusOK, res.StatusCode)
		})
	}
}

func TestShortenerHandlers_APIShortenURLBatch(t *testing.T) {
	type fields struct {
		shortenerService service.ShortenerService
		shortenedURLAddr string
		storage          storage.Storage
		contextTimeout   time.Duration
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	type ErrResponse struct {
		msg        string
		statusCode int
	}
	urlMap := make(map[string]model.ShortenedURL)
	userUrls := make([]model.ShortenedURL, 0)
	tests := []struct {
		name        string
		fields      fields
		args        args
		responseURL string
		wantErr     bool
		err         ErrResponse
	}{
		{
			name: "positive shorten url batch test",
			fields: fields{
				shortenerService: service.NewShortenerService(&MockStorage{urlMap, userUrls}, make(chan service.Task)),
				shortenedURLAddr: "http://localhost:8080",
				storage:          &MockStorage{urlMap, userUrls},
				contextTimeout:   time.Duration(2) * time.Second,
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost,
					"/api/shorten/batch",
					strings.NewReader(`
						[
							{
								"correlation_id": "1",
								"original_url": "https://google.com"
							},
							{
								"correlation_id": "2",
								"original_url": "https://ya.ru"
							},
							{
								"correlation_id": "3",	
								"original_url": "https://apple.com"
							}
						]
			`)),
			},
			responseURL: "http://localhost:8080/",
			wantErr:     false,
		},
		{
			name: "empty body",
			fields: fields{
				shortenerService: service.NewShortenerService(&MockStorage{urlMap, userUrls}, make(chan service.Task)),
				shortenedURLAddr: "http://localhost:8080",
				storage:          &MockStorage{urlMap, userUrls},
				contextTimeout:   time.Duration(2) * time.Second,
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost,
					"/api/shorten/batch",
					strings.NewReader(`[]`)),
			},
			responseURL: "http://localhost:8080/",
			wantErr:     true,
			err: ErrResponse{
				msg:        "Batch is empty\n",
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "context timeout",
			fields: fields{
				shortenerService: service.NewShortenerService(&MockStorage{urlMap, userUrls}, make(chan service.Task)),
				shortenedURLAddr: "http://localhost:8080",
				storage:          &MockStorage{urlMap, userUrls},
				contextTimeout:   time.Duration(0) * time.Second,
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost,
					"/api/shorten/batch",
					strings.NewReader(`
						[
							{
								"correlation_id": "1",
								"original_url": "https://google.com"
							},
							{
								"correlation_id": "2",
								"original_url": "https://ya.ru"
							},
							{
								"correlation_id": "3",	
								"original_url": "https://apple.com"
							}
						]
			`)),
			},
			responseURL: "http://localhost:8080/",
			wantErr:     true,
			err: ErrResponse{
				msg:        "Timeout exceeded\n",
				statusCode: http.StatusInternalServerError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sh := &ShortenerHandlers{
				shortenerService: tt.fields.shortenerService,
				shortenedURLAddr: tt.fields.shortenedURLAddr,
				storage:          tt.fields.storage,
				contextTimeout:   tt.fields.contextTimeout,
			}
			sh.APIShortenURLBatch(tt.args.w, tt.args.r)
			// assert response
			if !tt.wantErr {
				res := tt.args.w.(*httptest.ResponseRecorder).Result()
				assert.Equal(t, http.StatusCreated, res.StatusCode)
				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				err2 := res.Body.Close()
				require.NoError(t, err2)

				response := ExternalShortenedURLResponseDtoSlice{}
				err = response.UnmarshalJSON(body)
				if err != nil {
					assert.Fail(t, err.Error())
				}
				var dtos []ExternalShortenedURLResponseDto = response

				for i := 0; i < len(dtos); i++ {
					assert.Equal(t, 8, len(strings.Split(dtos[i].ShortURL, tt.responseURL)[1]))
				}
			} else {
				res := tt.args.w.(*httptest.ResponseRecorder).Result()
				assert.Equal(t, tt.err.statusCode, res.StatusCode)
				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				err2 := res.Body.Close()
				require.NoError(t, err2)
				assert.Equal(t, tt.err.msg, string(body))
			}
		})
	}
}

func TestShortenerHandlers_APIGetUserURLs(t *testing.T) {
	type fields struct {
		shortenerService service.ShortenerService
		shortenedURLAddr string
		storage          storage.Storage
		contextTimeout   time.Duration
	}
	type args struct {
		userUID uuid.UUID
		writer  http.ResponseWriter
		request *http.Request
	}
	type want struct {
		code     int
		response string
	}
	urlMap := make(map[string]model.ShortenedURL)
	userUrls := make([]model.ShortenedURL, 0)
	urlMap["dhKeUBD3"] = model.ShortenedURL{
		UUID:          uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		ShortURL:      "dhKeUBD3",
		OriginalURL:   "https://google.com",
		CorrelationID: sql.NullString{String: "correlation-id-1", Valid: true},
		DeletedFlag:   false,
	}
	urlMap["jnkGbkl2"] = model.ShortenedURL{
		UUID:          uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"),
		ShortURL:      "jnkGbkl2",
		OriginalURL:   "https://ya.ru",
		CorrelationID: sql.NullString{String: "correlation-id-2", Valid: true},
		DeletedFlag:   true,
	}
	userUrls = append(userUrls, urlMap["dhKeUBD3"], urlMap["jnkGbkl2"])

	storage := MockStorage{urlMap, userUrls}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "positive get user urls test",
			fields: fields{
				shortenerService: service.NewShortenerService(&storage, make(chan service.Task)),
				shortenedURLAddr: "http://localhost:8080",
				storage:          &storage,
				contextTimeout:   time.Duration(2) * time.Second,
			},
			args: args{
				userUID: uuid.MustParse("ec7325ca-a41a-49cc-8c21-f58d86385335"),
				writer:  httptest.NewRecorder(),
				request: httptest.NewRequest(http.MethodGet, "/api/user/urls", nil),
			},
			want: want{
				code: 200,
				response: `
					[
						{
							"short_url": "http://localhost:8080/dhKeUBD3",
							"original_url": "https://google.com"
						},
						{
							"short_url": "http://localhost:8080/jnkGbkl2",
							"original_url": "https://ya.ru"
						}
					]`,
			},
			wantErr: false,
		},
		{
			name: "context timeout",
			fields: fields{
				shortenerService: service.NewShortenerService(&storage, make(chan service.Task)),
				shortenedURLAddr: "http://localhost:8080",
				storage:          &storage,
				contextTimeout:   time.Duration(0) * time.Second,
			},
			args: args{
				userUID: uuid.MustParse("ec7325ca-a41a-49cc-8c21-f58d86385335"),
				writer:  httptest.NewRecorder(),
				request: httptest.NewRequest(http.MethodGet, "/api/user/urls", nil),
			},
			want: want{
				code:     http.StatusInternalServerError,
				response: "Timeout exceeded\n",
			},
			wantErr: true,
		},
		{
			name: "empty user uid",
			fields: fields{
				shortenerService: service.NewShortenerService(&storage, make(chan service.Task)),
				shortenedURLAddr: "http://localhost:8080",
				storage:          &storage,
				contextTimeout:   time.Duration(2) * time.Second,
			},
			args: args{
				userUID: uuid.Nil,
				writer:  httptest.NewRecorder(),
				request: httptest.NewRequest(http.MethodGet, "/api/user/urls", nil),
			},
			want: want{
				code:     http.StatusUnauthorized,
				response: "User is not authenticated\n",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sh := &ShortenerHandlers{
				shortenerService: tt.fields.shortenerService,
				shortenedURLAddr: tt.fields.shortenedURLAddr,
				storage:          tt.fields.storage,
				contextTimeout:   tt.fields.contextTimeout,
			}
			request := tt.args.request
			if tt.args.userUID != uuid.Nil {
				tt.args.request = request.WithContext(appContext.WithUserUID(request.Context(), &tt.args.userUID))
			}

			sh.APIGetUserURLs(tt.args.writer, tt.args.request)
			if !tt.wantErr {
				res := tt.args.writer.(*httptest.ResponseRecorder).Result()
				assert.Equal(t, http.StatusOK, res.StatusCode)
				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				err2 := res.Body.Close()
				require.NoError(t, err2)

				assert.JSONEqf(t, tt.want.response, string(body), "response body mismatch")
				assert.Equal(t, tt.want.code, res.StatusCode)
			} else {
				res := tt.args.writer.(*httptest.ResponseRecorder).Result()
				assert.Equal(t, tt.want.code, res.StatusCode)
				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				err2 := res.Body.Close()
				require.NoError(t, err2)
				assert.Equal(t, tt.want.response, string(body))
			}
		})
	}
}

func TestShortenerHandlers_APIDeleteUserURLs(t *testing.T) {
	urlMap := make(map[string]model.ShortenedURL)
	userUrls := make([]model.ShortenedURL, 0)

	s := MockStorage{urlMap, userUrls}
	type fields struct {
		shortenerService service.ShortenerService
		shortenedURLAddr string
		storage          storage.Storage
		contextTimeout   time.Duration
	}
	type args struct {
		userUID uuid.UUID
		writer  http.ResponseWriter
		request *http.Request
	}
	type want struct {
		code     int
		response string
	}
	taskChannel := make(chan service.Task, 100)
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "positive delete user urls test",
			fields: fields{
				shortenerService: service.NewShortenerService(&s, taskChannel),
				shortenedURLAddr: "http://localhost:8080",
				storage:          &s,
				contextTimeout:   time.Duration(2) * time.Second,
			},
			args: args{
				userUID: uuid.MustParse("ec7325ca-a41a-49cc-8c21-f58d86385335"),
				writer:  httptest.NewRecorder(),
				request: httptest.NewRequest(http.MethodDelete, "/api/user/urls", strings.NewReader(`
					["dhKeUBD3", "jnkGbkl2"]`)),
			},
			want: want{
				code: http.StatusAccepted,
			},
			wantErr: false,
		},
		{
			name: "context timeout",
			fields: fields{
				shortenerService: service.NewShortenerService(&s, taskChannel),
				shortenedURLAddr: "http://localhost:8080",
				storage:          &s,
				contextTimeout:   time.Duration(0) * time.Second,
			},
			args: args{
				userUID: uuid.MustParse("ec7325ca-a41a-49cc-8c21-f58d86385335"),
				writer:  httptest.NewRecorder(),
				request: httptest.NewRequest(http.MethodDelete, "/api/user/urls", strings.NewReader(`
					["dhKeUBD3", "jnkGbkl2"]`)),
			},
			want: want{
				code:     http.StatusInternalServerError,
				response: "Timeout exceeded\n",
			},
			wantErr: true,
		},
		{
			name: "empty user uid",
			fields: fields{
				shortenerService: service.NewShortenerService(&s, taskChannel),
				shortenedURLAddr: "http://localhost:8080",
				storage:          &s,
				contextTimeout:   time.Duration(2) * time.Second,
			},
			args: args{
				userUID: uuid.Nil,
				writer:  httptest.NewRecorder(),
				request: httptest.NewRequest(http.MethodDelete, "/api/user/urls", strings.NewReader(`
					["dhKeUBD3", "jnkGbkl2"]`)),
			},
			want: want{
				code:     http.StatusUnauthorized,
				response: "User is not authenticated\n",
			},
			wantErr: true,
		},
		{
			name: "empty list",
			fields: fields{
				shortenerService: service.NewShortenerService(&s, taskChannel),
				shortenedURLAddr: "http://localhost:8080",
				storage:          &s,
				contextTimeout:   time.Duration(2) * time.Second,
			},
			args: args{
				userUID: uuid.MustParse("ec7325ca-a41a-49cc-8c21-f58d86385335"),
				writer:  httptest.NewRecorder(),
				request: httptest.NewRequest(http.MethodDelete, "/api/user/urls", strings.NewReader(`
					[]`)),
			},
			want: want{
				code:     http.StatusBadRequest,
				response: "Batch is empty\n",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sh := &ShortenerHandlers{
				shortenerService: tt.fields.shortenerService,
				shortenedURLAddr: tt.fields.shortenedURLAddr,
				storage:          tt.fields.storage,
				contextTimeout:   tt.fields.contextTimeout,
			}
			request := tt.args.request
			if tt.args.userUID != uuid.Nil {
				tt.args.request = request.WithContext(appContext.WithUserUID(request.Context(), &tt.args.userUID))
			}

			sh.APIDeleteUserURLs(tt.args.writer, tt.args.request)
			// assert response
			if !tt.wantErr {
				res := tt.args.writer.(*httptest.ResponseRecorder).Result()
				err := res.Body.Close()
				require.NoError(t, err)
				assert.Equal(t, tt.want.code, res.StatusCode)
			} else {
				res := tt.args.writer.(*httptest.ResponseRecorder).Result()
				assert.Equal(t, tt.want.code, res.StatusCode)
				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				err2 := res.Body.Close()
				require.NoError(t, err2)
				assert.Equal(t, tt.want.response, string(body))
			}
		})
	}
}
