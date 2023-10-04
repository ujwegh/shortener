package handlers

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUrlShortener_ShortenUrl(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name    string
		method  string
		request string
		body    string
		want    want
	}{
		{
			name:    "positive shorten url test",
			request: "/",
			method:  http.MethodPost,
			body:    "https://google.com",
			want: want{
				code:        201,
				response:    "http://localhost:8080/",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:    "wrong http method",
			method:  http.MethodGet,
			request: "/",
			body:    "https://google.com",
			want: want{
				code:        400,
				contentType: "text/plain; charset=utf-8",
				response:    "Method not allowed\n",
			},
		},
		{
			name:    "empty request body",
			method:  http.MethodPost,
			request: "/",
			body:    "",
			want: want{
				code:        400,
				contentType: "text/plain; charset=utf-8",
				response:    "Url is empty\n",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.request, strings.NewReader(test.body))
			w := httptest.NewRecorder()
			var urlMap = make(map[string]string)
			us := &URLShortener{
				urlMap: urlMap,
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

func TestURLShortener_HandleShortenedURL(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	var targetURL = "https://google.com"
	key := "hdINdIoD"
	tests := []struct {
		urlMap  map[string]string
		name    string
		method  string
		request string
		body    string
		want    want
	}{
		{
			urlMap: map[string]string{
				key: targetURL,
			},
			name:    "positive shorten url test",
			request: "/" + key,
			method:  http.MethodGet,
			want: want{
				code:        307,
				contentType: "text/html; charset=utf-8",
				response:    targetURL,
			},
		},
		{
			urlMap: map[string]string{
				key: targetURL,
			},
			name:    "sent wrong key",
			request: "/" + "wrongKey",
			method:  http.MethodGet,
			want: want{
				code:        404,
				contentType: "text/plain; charset=utf-8",
				response:    "Shortened url not found\n",
			},
		},
		{
			urlMap: map[string]string{
				key: targetURL,
			},
			name:    "wrong method",
			request: "/" + key,
			method:  http.MethodPost,
			want: want{
				code:        400,
				contentType: "text/plain; charset=utf-8",
				response:    "Method not allowed\n",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.request, strings.NewReader(test.body))
			w := httptest.NewRecorder()
			us := &URLShortener{
				urlMap: test.urlMap,
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
