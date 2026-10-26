package dpop

import (
	"crypto/ecdsa"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

// Roundtripper used for fetching a DPoP token from the OIDC Provider. See https://datatracker.ietf.org/doc/html/rfc9449#name-dpop-access-token-request
type DPoPAccessTokenRoundTripper struct {
	Base    http.RoundTripper
	PrivKey *ecdsa.PrivateKey
}

func (ct DPoPAccessTokenRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {

	proof, err := GenerateDPoPProof(req, "", ct.PrivKey)

	if err != nil {
		return nil, fmt.Errorf("Failed to roundtrip DPoP token fetch request, %v, %v, %w", req.Method, req.RequestURI, err)
	}

	req2 := req.Clone(req.Context())
	req2.Header.Add("dpop", proof)

	return ct.Base.RoundTrip(req2)
}

// Roundtripper to add a DPoP proof header and 'Authorization : dpop <token>' header to a request. The token is obtained from an [oauth2.TokenSource] instance, see [oauth2.Token.AccessToken].
// See https://datatracker.ietf.org/doc/html/rfc9449#section-7.1
type DPoPResourceRoundTripper struct {
	Base        http.RoundTripper
	PrivKey     *ecdsa.PrivateKey
	TokenSource oauth2.TokenSource
}

func (ct DPoPResourceRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {

	token, err := ct.TokenSource.Token()

	if err != nil {
		return nil, fmt.Errorf("Failed to get token from oauth2 module:  %w", err)
	}

	proof, err := GenerateDPoPProof(req, token.AccessToken, ct.PrivKey)

	if err != nil {
		return nil, fmt.Errorf("Failed to roundtrip DPoP request, %v, %v, %w", req.Method, req.RequestURI, err)
	}

	req2 := req.Clone(req.Context())
	req2.Header.Add("dpop", proof)
	req2.Header.Add("Authorization", "Dpop "+token.AccessToken)

	return ct.Base.RoundTrip(req2)
}
