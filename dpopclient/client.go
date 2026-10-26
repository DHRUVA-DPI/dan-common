package dpopclient

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"net/http"

	"github.com/DHRUVA-DPI/dan-common/dpopclient/internal/dpop"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// DPoPClient.HttpClient is a [http.Client] configured with [http.RoundTrip] to add the DPoP token header and DPoP proof header.
// See https://datatracker.ietf.org/doc/html/rfc9449#section-7.1 . Use this client for communication with any DPoP token requiring server.
type DPoPClient struct {
	HttpClient  *http.Client
	tokenSource oauth2.TokenSource
}

// returns an instance of [DPoPClient]. oidcTokenUrl is the OIDC token endpoint. oidcClientId, oidcClientSecret are OIDC client credentials.
func NewDPoPClient(oidcTokenUrl string, oidcClientId string, oidcClientSecret string) (*DPoPClient, error) {
	ctx := context.Background()
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	if err != nil {
		return nil, fmt.Errorf("Failed to generate P256 keys : %w", err)
	}

	oidcConfig := clientcredentials.Config{ClientID: oidcClientId, ClientSecret: oidcClientSecret, TokenURL: oidcTokenUrl}

	dpopTokenClient := &http.Client{Transport: dpop.DPoPAccessTokenRoundTripper{Base: http.DefaultTransport, PrivKey: privKey}}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, dpopTokenClient)

	oidcTokenSource := oidcConfig.TokenSource(ctx)
	dpopResourceClient := &http.Client{Transport: dpop.DPoPResourceRoundTripper{Base: http.DefaultTransport, PrivKey: privKey, TokenSource: oidcTokenSource}}

	return &DPoPClient{HttpClient: dpopResourceClient, tokenSource: oidcTokenSource}, nil
}

// Test if OIDC config is correct and DPoP tokens can be obtained
func (client *DPoPClient) TestOIDCConfig() error {
	_, err := client.tokenSource.Token()
	return err
}
