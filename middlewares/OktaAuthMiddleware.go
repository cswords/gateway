package middlewares

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	verifier "github.com/okta/okta-jwt-verifier-golang"
)

// OIDCClient TODO
type OIDCClient struct {
	clientID string
	issuer   string
}

var tokens map[string]map[string]interface{} = make(map[string]map[string]interface{})

// NewOIDCClient TODO
func NewOIDCClient(clientID string, issuer string) *OIDCClient {
	c := OIDCClient{clientID: clientID, issuer: issuer}
	return &c
}

func (c *OIDCClient) verifyToken(t string) (*verifier.Jwt, error) {
	if claims, ok := tokens[t]; ok {
		expClaim := claims["exp"]
		if expClaim != nil {
			if exp, ok := expClaim.(int64); ok {
				expTime := time.Unix(exp, 0)
				log.Println("Token", t, "has been found from cache with expiration", expTime)
				if !expTime.After(time.Now()) {
					delete(tokens, t)
					return nil, fmt.Errorf("Token expeired: %s", t)
				}
			}
		}
		result := verifier.Jwt{Claims: claims}
		return &result, nil
	}
	log.Println("Token", t, "was not found and will be validated")

	tv := map[string]string{}
	tv["aud"] = "api://default"
	tv["cid"] = c.clientID

	jv := verifier.JwtVerifier{
		Issuer:           c.issuer,
		ClaimsToValidate: tv,
	}

	result, err := jv.New().VerifyAccessToken(t)

	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}

	if result != nil {
		tokens[t] = result.Claims
		return result, nil
	}

	return nil, fmt.Errorf("token could not be verified: %s", t)
}

// NewOktaAuthMiddleware TODO
func NewOktaAuthMiddleware(c *OIDCClient) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				jwt, err := c.verifyToken(token)
				if err != nil {
					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}
				for k, v := range jwt.Claims {
					key := "x-okta-claim-" + k
					value := fmt.Sprintf("%v", v)
					r.Header.Set(key, value)
				}
				next.ServeHTTP(w, r)
				return
			}
			http.Error(w, "Okta token cannot be found in the header of this request", http.StatusUnauthorized)
		})
	}
}
