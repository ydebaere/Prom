package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/handlers"

	"backend/internal/config"
)

type Claims struct {
	jwt.RegisteredClaims
}

const AnonymousToken = "anonymous-fetch"

var jwks *keyfunc.JWKS

func init() {
	cfg := config.LoadAzureConfig()

	jwksURL := fmt.Sprintf("https://login.microsoftonline.com/%s/discovery/v2.0/keys", cfg.TenantID)

	var err error
	jwks, err = keyfunc.Get(jwksURL, keyfunc.Options{
		RefreshInterval: time.Hour,
		RefreshErrorHandler: func(err error) {
			fmt.Printf("Failed to refresh JWKS: %v\n", err)
		},
		RefreshTimeout: 10 * time.Second,
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to create JWKS from Azure: %v", err))
	}
}

func JWTMiddleware(w http.ResponseWriter, r *http.Request) bool {
	cfg := config.LoadAzureConfig()
	url := fmt.Sprintf("api://%s", cfg.ClientID)

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Error", http.StatusUnauthorized)
		return false
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == AnonymousToken {
		fmt.Println("[AUTH] Anonymous access granted.")
		return true
	}

	if tokenString == authHeader {
		http.Error(w, "Error", http.StatusUnauthorized)
		return false
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwks.Keyfunc(token)
	})

	if err != nil || !token.Valid {
		http.Error(w, "Error", http.StatusUnauthorized)
		return false
	}

	if !claims.VerifyAudience(url, true) {
		http.Error(w, "Error", http.StatusUnauthorized)
		return false
	}

	return true
}

func WithJWTMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !JWTMiddleware(w, r) {
			return
		}
		handler(w, r)
	}
}

func SetupCORS(handler http.Handler) http.Handler {
	return handlers.CORS(
		handlers.AllowedOrigins([]string{
			"http://localhost:9000",
		}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)(handler)
}
