package storage

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	appErrors "github.com/ujwegh/shortener/internal/app/errors"
	"github.com/ujwegh/shortener/internal/app/model"
	"reflect"
	"testing"
)

const initDB = `
CREATE TABLE IF NOT EXISTS shortened_urls
(
    uuid TEXT PRIMARY KEY,
    short_url TEXT UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    correlation_id TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS shortened_urls_correlation_id_idx ON shortened_urls (correlation_id)
    WHERE correlation_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS original_urls_unique_idx ON shortened_urls (original_url);
`

func setupInMemoryDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite3", "file:memdb1?mode=memory&cache=shared")

	if err != nil {
		t.Fatalf("could not create in-memory db: %v", err)
	}

	_, err = db.Exec(initDB)
	if err != nil {
		t.Fatalf("could not create table: %v", err)
	}

	return db
}

func TestDBStorage_Ping(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	type fields struct {
		db *sqlx.DB
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "ping success",
			fields:  fields{db: db},
			args:    args{ctx: context.Background()},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &DBStorage{
				db: tt.fields.db,
			}
			if err := storage.Ping(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Ping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDBStorage_ReadShortenedURL(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	dbTestData := `
DELETE FROM shortened_urls;
INSERT INTO shortened_urls (uuid, short_url, original_url, correlation_id) 
VALUES ('c12ff52b-970a-479c-bd45-1c6043c98736', 'abxW9ymI', 'https://ya.ru', 'correlation1'),
       ('cb280de3-c5ba-4fab-92d9-30bd72282afc', 'E9M9zboP', 'https://google.com', null);
`

	type fields struct {
		db *sqlx.DB
	}
	type args struct {
		ctx context.Context
		url string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.ShortenedURL
		wantErr bool
	}{
		{
			name:   "read by short url - success",
			fields: fields{db: db},
			args:   args{ctx: context.Background(), url: "abxW9ymI"},
			want: &model.ShortenedURL{
				UUID:        uuid.MustParse("c12ff52b-970a-479c-bd45-1c6043c98736"),
				ShortURL:    "abxW9ymI",
				OriginalURL: "https://ya.ru",
				CorrelationID: sql.NullString{
					String: "correlation1",
					Valid:  true,
				},
			},
			wantErr: false,
		},
		{
			name:   "read by original url - success",
			fields: fields{db: db},
			args:   args{ctx: context.Background(), url: "https://google.com"},
			want: &model.ShortenedURL{
				UUID:        uuid.MustParse("cb280de3-c5ba-4fab-92d9-30bd72282afc"),
				ShortURL:    "E9M9zboP",
				OriginalURL: "https://google.com",
				CorrelationID: sql.NullString{
					String: "",
					Valid:  false,
				},
			},
			wantErr: false,
		},
		{
			name:    "read by non-existent url - error",
			fields:  fields{db: db},
			args:    args{ctx: context.Background(), url: "https://non-existent.com"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := db.Exec(dbTestData)
			if err != nil {
				t.Fatalf("could not insert test data: %v", err)
			}
			storage := &DBStorage{
				db: tt.fields.db,
			}
			got, err := storage.ReadShortenedURL(tt.args.ctx, tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadShortenedURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadShortenedURL() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBStorage_WriteBatchShortenedURLSlice(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	type fields struct {
		db *sqlx.DB
	}
	type args struct {
		ctx       context.Context
		urlsSlice []model.ShortenedURL
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []model.ShortenedURL
		wantErr bool
	}{
		{
			name:   "write batch - success",
			fields: fields{db: db},
			args: args{
				ctx: context.Background(),
				urlsSlice: []model.ShortenedURL{
					{
						UUID:        uuid.MustParse("c12ff52b-970a-479c-bd45-1c6043c98736"),
						ShortURL:    "abxW9ymI",
						OriginalURL: "https://ya.ru",
						CorrelationID: sql.NullString{
							String: "correlation1",
							Valid:  true,
						},
					},
					{
						UUID:        uuid.MustParse("cb280de3-c5ba-4fab-92d9-30bd72282afc"),
						ShortURL:    "E9M9zboP",
						OriginalURL: "https://google.com",
						CorrelationID: sql.NullString{
							String: "correlation2",
							Valid:  true,
						},
					},
				},
			},
			want: []model.ShortenedURL{
				{
					UUID:        uuid.MustParse("c12ff52b-970a-479c-bd45-1c6043c98736"),
					ShortURL:    "abxW9ymI",
					OriginalURL: "https://ya.ru",
					CorrelationID: sql.NullString{
						String: "correlation1",
						Valid:  true,
					},
				},
				{
					UUID:        uuid.MustParse("cb280de3-c5ba-4fab-92d9-30bd72282afc"),
					ShortURL:    "E9M9zboP",
					OriginalURL: "https://google.com",
					CorrelationID: sql.NullString{
						String: "correlation2",
						Valid:  true,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &DBStorage{
				db: tt.fields.db,
			}
			if err := storage.WriteBatchShortenedURLSlice(tt.args.ctx, tt.args.urlsSlice); (err != nil) != tt.wantErr {
				t.Errorf("WriteBatchShortenedURLSlice() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.args.urlsSlice, tt.want) {
				t.Errorf("WriteBatchShortenedURLSlice() got = %v, want %v", tt.args.urlsSlice, tt.want)
			}

			if tt.wantErr == false {
				var count int64
				err2 := db.QueryRow("SELECT COUNT(*) FROM shortened_urls").Scan(&count)
				if err2 != nil {
					t.Errorf("could not get count: %v", err2)
				}
				if count != 2 {
					t.Errorf("expected 2 rows, got %d", count)
				}
			}
		})
	}
}

func TestDBStorage_WriteShortenedURL(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	dbTestData := `
DELETE FROM shortened_urls;
INSERT INTO shortened_urls (uuid, short_url, original_url, correlation_id) 
VALUES ('c12ff52b-970a-479c-bd45-1c6043c98736', 'abxW9ymI', 'https://ya.ru', 'correlation1'),
       ('cb280de3-c5ba-4fab-92d9-30bd72282afc', 'E9M9zboP', 'https://google.com', null);
`

	type fields struct {
		db *sqlx.DB
	}
	type args struct {
		ctx          context.Context
		shortenedURL *model.ShortenedURL
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "write - success",
			fields: fields{db: db},
			args: args{
				ctx: context.Background(),
				shortenedURL: &model.ShortenedURL{
					UUID:        uuid.MustParse("f2c7c737-b70d-49cb-a0eb-079e10e8ed29"),
					ShortURL:    "BDurKLrm",
					OriginalURL: "https://yandex.ru",
					CorrelationID: sql.NullString{
						String: "",
						Valid:  false,
					},
				},
			},
			wantErr: false,
		},
		{
			name:   "write - unique violation",
			fields: fields{db: db},
			args: args{
				ctx: context.Background(),
				shortenedURL: &model.ShortenedURL{
					UUID:        uuid.MustParse("c12ff52b-970a-479c-bd45-1c6043c98736"),
					ShortURL:    "abxW9ymI",
					OriginalURL: "https://ya.ru",
					CorrelationID: sql.NullString{
						String: "",
						Valid:  false,
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := db.Exec(dbTestData)
			if err != nil {
				t.Fatalf("could not insert test data: %v", err)
			}

			storage := &DBStorage{
				db: tt.fields.db,
			}
			err = storage.WriteShortenedURL(tt.args.ctx, tt.args.shortenedURL)
			if err != nil {
				if tt.wantErr == true {
					target := &appErrors.ShortenerError{}
					assert.True(t, true, errors.As(err, target) && target.Msg() == "unique violation")
				} else {
					t.Errorf("WriteShortenedURL() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			if tt.wantErr == false {
				var count int64
				err2 := db.QueryRow("SELECT COUNT(*) FROM shortened_urls").Scan(&count)
				if err2 != nil {
					t.Errorf("could not get count: %v", err2)
				}
				if count != 3 {
					t.Errorf("expected 2 row, got %d", count)
				}
			}

		})
	}
}
