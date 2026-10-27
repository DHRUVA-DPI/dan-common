package dpop

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/AxisCommunications/go-dpop"
	"github.com/golang-jwt/jwt/v5"
)

// generates a DPoP proof given an [http.Request]. If token is passed, then the `ath` claim is also added to the DPoP proof.
// The `ath` claim is necessary for requests including the `Authorization: Dpop ...` header
func GenerateDPoPProof(r *http.Request, token string, privateKey *ecdsa.PrivateKey) (string, error) {

	// making copy of the URL to remove query and fragment, since `htu` URL must not have these
	// use .Clone method when moving to go 1.27
	urlCopy := *(r.URL)
	urlCopy.RawQuery = ""
	urlCopy.Fragment = ""

	claims := dpop.ProofTokenClaims{
		RegisteredClaims: &jwt.RegisteredClaims{
			ID:       rand.Text(),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
		Method: dpop.HTTPVerb(r.Method),
		URL:    urlCopy.String(),
	}

	if token != "" {
		h := sha256.New()
		_, err := h.Write([]byte(token))

		if err != nil {
			return "", fmt.Errorf("Failed to create `ath` of token : %w", err)
		}
		ath := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
		claims.AccessTokenHash = ath
	}

	proof, err := dpop.Create(jwt.SigningMethodES256, &claims, privateKey)
	if err != nil {
		return "", fmt.Errorf("Failed to create DPoP proof for %v, %v : %w", claims.Method, claims.URL, err)
	}

	return proof, nil
}
