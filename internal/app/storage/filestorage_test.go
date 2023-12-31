package storage

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/ujwegh/shortener/internal/app/config"
	"github.com/ujwegh/shortener/internal/app/model"
	"os"
	"reflect"
	"testing"
)

func TestFileStorage_ReadShortenedURL(t *testing.T) {
	// prepare test data
	testShortenedURLsFileName := "/tmp/shortened-urls-test.json"
	file, err := os.OpenFile(testShortenedURLsFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		t.Error(err)
	}
	encoder := json.NewEncoder(file)
	uid := uuid.New()
	testUserURLsFileName := "/tmp/user-urls-test.json"

	err = encoder.Encode(&model.ShortenedURL{
		UUID:        uid,
		ShortURL:    "edVPg3ks",
		OriginalURL: "http://ya.ru",
	})
	if err != nil {
		t.Error(err)
	}

	appConfig := config.AppConfig{
		ShortenedURLsFilePath: testShortenedURLsFileName,
		UserURLsFilePath:      testUserURLsFileName,
	}
	type fields struct {
		cfg    config.AppConfig
		urlMap map[string]string
	}
	type args struct {
		shortURL string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.ShortenedURL
		wantErr bool
	}{
		{
			name: "positive read shortened url test",
			fields: fields{
				cfg: appConfig,
			},
			args: args{
				shortURL: "edVPg3ks",
			},
			want: &model.ShortenedURL{
				UUID:        uid,
				ShortURL:    "edVPg3ks",
				OriginalURL: "http://ya.ru",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fss := NewFileStorage(tt.fields.cfg)
			got, err := fss.ReadShortenedURL(context.Background(), tt.args.shortURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadShortenedURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadShortenedURL() got = %v, want %v", got, tt.want)
			}
		})
	}
	defer func() {
		if err := os.Remove(testShortenedURLsFileName); err != nil {
			t.Error(err)
		}
	}()
}

func TestFileStorage_WriteShortenedURL(t *testing.T) {
	// prepare test data
	testShortenedURLsFileName := "/tmp/shortened-urls-test.json"
	_, err := os.OpenFile(testShortenedURLsFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		t.Error(err)
	}
	testUserURLsFileName := "/tmp/user-urls-test.json"
	appConfig := config.AppConfig{
		ShortenedURLsFilePath: testShortenedURLsFileName,
		UserURLsFilePath:      testUserURLsFileName,
	}

	type fields struct {
		cfg    config.AppConfig
		urlMap map[string]string
	}
	type args struct {
		shortenedURL *model.ShortenedURL
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "positive write shortened url test",
			fields: fields{
				cfg: appConfig,
			},
			args: args{
				shortenedURL: &model.ShortenedURL{
					UUID:        uuid.New(),
					ShortURL:    "edJkl5jj",
					OriginalURL: "http://ya.com",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fss := NewFileStorage(tt.fields.cfg)
			if err := fss.WriteShortenedURL(context.Background(), tt.args.shortenedURL); (err != nil) != tt.wantErr {
				t.Errorf("WriteShortenedURL() error = %v, wantErr %v", err, tt.wantErr)
			}

			// check if the data was written to the file
			file, err := os.OpenFile(testShortenedURLsFileName, os.O_RDONLY, 0666)
			if err != nil {
				t.Error(err)
			}
			decoder := json.NewDecoder(file)
			var shortenedURL model.ShortenedURL
			err = decoder.Decode(&shortenedURL)
			if err != nil {
				t.Error(err)
			}
			if shortenedURL.ShortURL != tt.args.shortenedURL.ShortURL {
				t.Errorf("WriteShortenedURL() got = %v, want %v", shortenedURL.ShortURL, tt.args.shortenedURL.ShortURL)
			}
			// assert internal shortURLMap not empty
			if len(fss.shortURLMap) == 0 {
				t.Errorf("WriteShortenedURL() got = %v, want %v", len(fss.shortURLMap), 1)
			}
		})
	}
	defer func() {
		if err := os.Remove(testShortenedURLsFileName); err != nil {
			t.Error(err)
		}
	}()
}

func TestFileStorage_readAllShortenedURLs(t *testing.T) {
	// prepare test data
	testShortenedURLsFileName := "/tmp/shortened-urls-test.json"
	file, err := os.OpenFile(testShortenedURLsFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		t.Error(err)
	}
	encoder := json.NewEncoder(file)
	uid := uuid.New()
	err = encoder.Encode(&model.ShortenedURL{
		UUID:        uid,
		ShortURL:    "edVPg3ks",
		OriginalURL: "http://ya.ru",
	})
	if err != nil {
		t.Error(err)
	}
	testUserURLsFileName := "/tmp/user-urls-test.json"
	appConfig := config.AppConfig{
		ShortenedURLsFilePath: testShortenedURLsFileName,
		UserURLsFilePath:      testUserURLsFileName,
	}

	type fields struct {
		cfg    config.AppConfig
		urlMap map[string]model.ShortenedURL
	}
	tests := []struct {
		name    string
		fields  fields
		want    []model.ShortenedURL
		wantErr bool
	}{
		{
			name: "positive read all shortened urls test",
			fields: fields{
				cfg: appConfig,
			},
			want: []model.ShortenedURL{
				{
					UUID:        uid,
					ShortURL:    "edVPg3ks",
					OriginalURL: "http://ya.ru",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fss := NewFileStorage(tt.fields.cfg)
			got, err := fss.readAllShortenedURLs()
			if (err != nil) != tt.wantErr {
				t.Errorf("readAllShortenedURLs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readAllShortenedURLs() got = %v, want %v", got, tt.want)
			}
		})
	}
}
