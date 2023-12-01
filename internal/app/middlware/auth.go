package middlware

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/ujwegh/shortener/internal/app/context"
	appErrors "github.com/ujwegh/shortener/internal/app/errors"
	"github.com/ujwegh/shortener/internal/app/logger"
	"github.com/ujwegh/shortener/internal/app/service"
	"go.uber.org/zap"
	"net/http"
)

type AuthMiddleware struct {
	tokenService service.TokenService
}

func NewAuthMiddleware(tokenService service.TokenService) AuthMiddleware {
	return AuthMiddleware{
		tokenService: tokenService,
	}
}

func (am *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := readCookie(r, CookieSessionToken)
		if err != nil {
			am.createNewToken(w, r, next)
			return
		}
		uidString, err := am.tokenService.GetUserUID(token)
		if err != nil {
			target := &appErrors.ShortenerError{}
			if errors.As(err, target) {
				logger.Log.Warn("failed to get userUID: " + target.Msg())
				am.createNewToken(w, r, next)
				return
			} else {
				fmt.Printf("failed to get userUID: %s", err)
				http.Error(w, "Something went wrong.", http.StatusInternalServerError)
				return
			}
		}
		if uidString == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		userUID, err := uuid.Parse(uidString)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		token, err = am.tokenService.GenerateToken(&userUID)
		if err != nil {
			logger.Log.Error("failed to generate token", zap.Error(err))
			http.Error(w, "Something went wrong.", http.StatusInternalServerError)
			return
		}
		setCookie(w, CookieSessionToken, token)
		next.ServeHTTP(w, setUser(r, &userUID))
	})
}

func (am *AuthMiddleware) createNewToken(w http.ResponseWriter, r *http.Request, next http.Handler) {
	userUID := uuid.New()
	token, err := am.tokenService.GenerateToken(&userUID)
	if err != nil {
		fmt.Printf("failed to generate token: %s", err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	setCookie(w, CookieSessionToken, token)
	next.ServeHTTP(w, setUser(r, &userUID))
}

func setUser(r *http.Request, userUID *uuid.UUID) *http.Request {
	ctx := r.Context()
	ctx = context.WithUserUID(ctx, userUID)
	return r.WithContext(ctx)
}
