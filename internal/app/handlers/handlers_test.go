package handlers

import (
	"net/http"
	"testing"
)

func TestUrlShortener_ShortenUrl(t *testing.T) {
	type fields struct {
		urlMap map[string]string
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			us := &UrlShortener{
				urlMap: tt.fields.urlMap,
			}
			us.ShortenUrl(tt.args.w, tt.args.r)
		})
	}
}
