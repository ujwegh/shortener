package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	appContext "github.com/ujwegh/shortener/internal/app/context"
	appErrors "github.com/ujwegh/shortener/internal/app/errors"
	"github.com/ujwegh/shortener/internal/app/logger"
	"github.com/ujwegh/shortener/internal/app/model"
	"github.com/ujwegh/shortener/internal/app/service"
	"github.com/ujwegh/shortener/internal/app/storage"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

type (
	ShortenerHandlers struct {
		shortenerService service.ShortenerService
		shortenedURLAddr string
		storage          storage.Storage
		contextTimeout   time.Duration
	}
	//easyjson:json
	ShortenRequestDto struct {
		URL string `json:"url"`
	}
	//easyjson:json
	ShortenResponseDto struct {
		Result string `json:"result"`
	}
	//easyjson:json
	ExternalShortenedURLRequestDto struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}
	//easyjson:json
	ExternalShortenedURLResponseDto struct {
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
	}
	//easyjson:json
	ExternalShortenedURLRequestDtoSlice []ExternalShortenedURLRequestDto
	//easyjson:json
	ExternalShortenedURLResponseDtoSlice []ExternalShortenedURLResponseDto
	//easyjson:json
	UserURLDto struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}
	//easyjson:json
	UserURLDtoSlice []UserURLDto

	//easyjson:json
	DeleteUserURLsDto []string
)

const errMsgCreateShortURL = "Unable to create shortened URL"
const errMsgEnableReadBody = "Unable to read body"

func NewShortenerHandlers(shortenedURLAddr string, contextTimeout int, service service.ShortenerService, storage storage.Storage) *ShortenerHandlers {
	return &ShortenerHandlers{
		shortenerService: service,
		storage:          storage,
		shortenedURLAddr: shortenedURLAddr,
		contextTimeout:   time.Duration(contextTimeout) * time.Second,
	}
}

func (sh *ShortenerHandlers) ShortenURL(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), sh.contextTimeout)
	defer cancel()

	userUID := appContext.UserUID(r.Context())
	if userUID == nil {
		http.Error(w, "User is not authenticated", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, errMsgEnableReadBody, http.StatusBadRequest)
		return
	}
	originalURL := string(body)
	if originalURL == "" {
		http.Error(w, "Url is empty", http.StatusBadRequest)
		return
	}
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	shortenedURL, err := sh.shortenerService.CreateShortenedURL(ctx, userUID, originalURL)
	shortenedURL, hasError := sh.checkCreateShortenedURLError(ctx, w, err, shortenedURL, originalURL)
	if hasError {
		return
	}

	if contextHasError(w, ctx) {
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s/%s", sh.shortenedURLAddr, shortenedURL.ShortURL)
}

func (sh *ShortenerHandlers) APIShortenURL(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), sh.contextTimeout)
	defer cancel()

	userUID := appContext.UserUID(r.Context())
	if userUID == nil {
		http.Error(w, "User is not authenticated", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, errMsgEnableReadBody, http.StatusBadRequest)
		return
	}
	request := ShortenRequestDto{}
	err = request.UnmarshalJSON(body)
	if err != nil {
		http.Error(w, "Unable to parse body", http.StatusBadRequest)
		return
	}
	originalURL := request.URL
	if originalURL == "" {
		http.Error(w, "URL is empty", http.StatusBadRequest)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	shortenedURL, err := sh.shortenerService.CreateShortenedURL(ctx, userUID, originalURL)
	shortenedURL, hasError := sh.checkCreateShortenedURLError(ctx, w, err, shortenedURL, originalURL)
	if hasError {
		return
	}

	response := &ShortenResponseDto{Result: fmt.Sprintf("%s/%s", sh.shortenedURLAddr, shortenedURL.ShortURL)}
	rawBytes, err := response.MarshalJSON()
	if err != nil {
		http.Error(w, "Unable to marshal response", http.StatusInternalServerError)
		return
	}
	if contextHasError(w, ctx) {
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s", rawBytes)
}

func (sh *ShortenerHandlers) checkCreateShortenedURLError(ctx context.Context, w http.ResponseWriter, err error, shortenedURL *model.ShortenedURL, originalURL string) (*model.ShortenedURL, bool) {
	shortenerError := appErrors.ShortenerError{}
	if err != nil && errors.As(err, &shortenerError) && shortenerError.Msg() == "unique violation" {
		shortenedURL, err = sh.shortenerService.GetShortenedURL(ctx, originalURL)
		if err != nil {
			logger.Log.Error(errMsgCreateShortURL, zap.Error(err))
			http.Error(w, errMsgCreateShortURL, http.StatusInternalServerError)
			return nil, true
		}
		w.WriteHeader(http.StatusConflict)
	} else if err != nil {
		logger.Log.Error(errMsgCreateShortURL, zap.Error(err))
		http.Error(w, errMsgCreateShortURL, http.StatusInternalServerError)
		return nil, true
	}
	return shortenedURL, false
}

func (sh *ShortenerHandlers) HandleShortenedURL(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), sh.contextTimeout)
	defer cancel()
	shortKey := chi.URLParam(r, "id")
	shortenedURL, err := sh.shortenerService.GetShortenedURL(ctx, shortKey)
	if err != nil {
		http.Error(w, "Unable to get shortened URL", http.StatusInternalServerError)
		return
	}
	originalURL := shortenedURL.OriginalURL
	if originalURL == "" {
		http.Error(w, "Shortened url not found", http.StatusNotFound)
		return
	}

	if shortenedURL.DeletedFlag {
		w.WriteHeader(http.StatusGone)
		return
	}

	if contextHasError(w, ctx) {
		return
	}
	w.Header().Add("Location", originalURL)
	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

func (sh *ShortenerHandlers) Ping(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), sh.contextTimeout)
	defer cancel()
	err := sh.storage.Ping(ctx)
	if err != nil {
		http.Error(w, "Unable to ping storage", http.StatusInternalServerError)
		return
	}

	if contextHasError(w, ctx) {
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (sh *ShortenerHandlers) APIShortenURLBatch(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), sh.contextTimeout)
	defer cancel()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, errMsgEnableReadBody, http.StatusBadRequest)
		return
	}
	request := ExternalShortenedURLRequestDtoSlice{}
	err = request.UnmarshalJSON(body)
	if err != nil {
		http.Error(w, "Unable to parse body", http.StatusBadRequest)
		return
	}
	var dtos []ExternalShortenedURLRequestDto = request
	if len(dtos) == 0 {
		http.Error(w, "Batch is empty", http.StatusBadRequest)
		return
	}
	urls := mapExternalRequestToShortenedURL(dtos)
	shortenedURLs, err := sh.shortenerService.BatchCreateShortenedURLs(ctx, *urls)
	if err != nil {
		http.Error(w, "Unable to batch insert shortened URLs", http.StatusInternalServerError)
		return
	}
	response := mapShortenedURLToExternalResponse(sh, *shortenedURLs)
	rawBytes, err := response.MarshalJSON()
	if err != nil {
		http.Error(w, "Unable to marshal response", http.StatusInternalServerError)
		return
	}
	if contextHasError(w, ctx) {
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s", rawBytes)
}

func (sh *ShortenerHandlers) APIGetUserURLs(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), sh.contextTimeout)
	defer cancel()
	userUID := appContext.UserUID(request.Context())
	if userUID == nil {
		http.Error(writer, "User is not authenticated", http.StatusUnauthorized)
		return
	}
	shortenedURLs, err := sh.shortenerService.GetUserShortenedURLs(ctx, userUID)
	if err != nil {
		http.Error(writer, "Unable to get user URLs", http.StatusInternalServerError)
		return
	}
	if len(*shortenedURLs) == 0 {
		writer.WriteHeader(http.StatusNoContent)
		return
	}
	response := mapShortenedURLToUserURLDtoSlice(sh, *shortenedURLs)
	rawBytes, err := response.MarshalJSON()
	if err != nil {
		http.Error(writer, "Unable to marshal response", http.StatusInternalServerError)
		return
	}
	if contextHasError(writer, ctx) {
		return
	}
	writer.Header().Add("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	fmt.Fprintf(writer, "%s", rawBytes)
}

func (sh *ShortenerHandlers) APIDeleteUserURLs(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), sh.contextTimeout)
	defer cancel()
	userUID := appContext.UserUID(request.Context())
	if userUID == nil {
		http.Error(writer, "User is not authenticated", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, errMsgEnableReadBody, http.StatusBadRequest)
		return
	}

	requestBody := DeleteUserURLsDto{}
	err = requestBody.UnmarshalJSON(body)
	if err != nil {
		http.Error(writer, "Unable to parse body", http.StatusBadRequest)
		return
	}
	var shortURLKeys []string = requestBody
	if len(shortURLKeys) == 0 {
		http.Error(writer, "Batch is empty", http.StatusBadRequest)
		return
	}

	err = sh.shortenerService.DeleteUserShortenedURLs(ctx, userUID, shortURLKeys)

	if err != nil {
		http.Error(writer, "Unable to delete user URLs", http.StatusInternalServerError)
		return
	}

	if contextHasError(writer, ctx) {
		return
	}
	writer.WriteHeader(http.StatusAccepted)
}

func contextHasError(w http.ResponseWriter, ctx context.Context) bool {
	if err := ctx.Err(); err != nil {
		var errMsg string
		var errCode int

		switch err {
		case context.Canceled:
			errMsg, errCode = "Request canceled", http.StatusInternalServerError
		case context.DeadlineExceeded:
			errMsg, errCode = "Timeout exceeded", http.StatusInternalServerError
		default:
			return false
		}

		http.Error(w, errMsg, errCode)
		return true
	}
	return false
}

func mapShortenedURLToExternalResponse(sh *ShortenerHandlers, slice []model.ShortenedURL) ExternalShortenedURLResponseDtoSlice {
	var responseSlice []ExternalShortenedURLResponseDto
	for _, item := range slice {
		responseItem := ExternalShortenedURLResponseDto{
			CorrelationID: item.CorrelationID.String,
			ShortURL:      fmt.Sprintf("%s/%s", sh.shortenedURLAddr, item.ShortURL),
		}
		responseSlice = append(responseSlice, responseItem)
	}
	return responseSlice
}
func mapShortenedURLToUserURLDtoSlice(sh *ShortenerHandlers, slice []model.ShortenedURL) UserURLDtoSlice {
	var responseSlice []UserURLDto
	for _, item := range slice {
		responseItem := UserURLDto{
			OriginalURL: item.OriginalURL,
			ShortURL:    fmt.Sprintf("%s/%s", sh.shortenedURLAddr, item.ShortURL),
		}
		responseSlice = append(responseSlice, responseItem)
	}
	return responseSlice
}

func mapExternalRequestToShortenedURL(slice []ExternalShortenedURLRequestDto) *[]model.ShortenedURL {
	var shortenedURLs []model.ShortenedURL
	for _, item := range slice {
		shortenedURL := model.ShortenedURL{
			CorrelationID: sql.NullString{
				String: item.CorrelationID,
				Valid:  true,
			},
			OriginalURL: item.OriginalURL,
		}
		shortenedURLs = append(shortenedURLs, shortenedURL)
	}
	return &shortenedURLs
}
