package dpop

import (
	"crypto/ecdsa"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

// Roundtrip logic copied to some extent from golang oauth2 package
// https://github.com/golang/oauth2/blob/c624b89dadc3221560b7345c090bbe69e90808ee/transport.go

// Roundtripper used for fetching a DPoP token from the OIDC Provider.
// See https://datatracker.ietf.org/doc/html/rfc9449#name-dpop-access-token-request
type DPoPAccessTokenRoundTripper struct {
	Base    http.RoundTripper
	PrivKey *ecdsa.PrivateKey
}

func (ct DPoPAccessTokenRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {

	// part of oauth2/transport.go - Roundtrip must close request body always
	reqBodyClosed := false
	if req.Body != nil {
		defer func() {
			if !reqBodyClosed {
				req.Body.Close()
			}
		}()
	}

	proof, err := GenerateDPoPProof(req, "", ct.PrivKey)

	if err != nil {
		return nil, fmt.Errorf("Failed to roundtrip DPoP token fetch request, %v, %v, %w", req.Method, req.URL.String(), err)
	}

	req2 := req.Clone(req.Context())
	req2.Header.Set("DPoP", proof)

	// part of oauth2/transport.go
	// req.Body is assumed to be closed by the base RoundTripper.
	reqBodyClosed = true
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

	// part of oauth2/transport.go - Roundtrip must close request body always
	reqBodyClosed := false
	if req.Body != nil {
		defer func() {
			if !reqBodyClosed {
				req.Body.Close()
			}
		}()
	}

	token, err := ct.TokenSource.Token()

	if err != nil {
		return nil, fmt.Errorf("Failed to get token from oauth2 module:  %w", err)
	}

	proof, err := GenerateDPoPProof(req, token.AccessToken, ct.PrivKey)

	if err != nil {
		return nil, fmt.Errorf("Failed to roundtrip DPoP request, %v, %v, %w", req.Method, req.URL.String(), err)
	}

	req2 := req.Clone(req.Context())
	req2.Header.Set("DPoP", proof)
	req2.Header.Set("Authorization", "DPoP "+token.AccessToken)

	// part of oauth2/transport.go
	// req.Body is assumed to be closed by the base RoundTripper.
	reqBodyClosed = true
	return ct.Base.RoundTrip(req2)
}
