package uaa

import (
	"context"
	"net/http"
	"net/url"

	"github.com/cloudfoundry-community/go-uaa/passwordcredentials"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// API is a client to the UAA API.
type API struct {
	AuthenticatedClient   *http.Client
	UnauthenticatedClient *http.Client
	TargetURL             *url.URL
	SkipSSLValidation     bool
	Verbose               bool
	ZoneID                string
}

// NewWithClientCredentials builds an API that uses the client credentials grant
// to get a token for use with the UAA API.
func NewWithClientCredentials(target string, zoneID string, clientID string, clientSecret string) (*API, error) {
	url, err := BuildTargetURL(target)
	if err != nil {
		return nil, err
	}

	tokenURL := urlWithPath(*url, "/oauth/token")
	c := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     tokenURL.String(),
	}
	client := &http.Client{Transport: http.DefaultTransport}
	return &API{
		UnauthenticatedClient: client,
		AuthenticatedClient:   c.Client(context.WithValue(context.Background(), oauth2.HTTPClient, client)),
		TargetURL:             url,
		ZoneID:                zoneID,
	}, nil
}

// NewWithPasswordCredentials builds an API that uses the password credentials
// grant to get a token for use with the UAA API.
func NewWithPasswordCredentials(target string, zoneID string, clientID string, clientSecret string, username string, password string) (*API, error) {
	url, err := BuildTargetURL(target)
	if err != nil {
		return nil, err
	}

	tokenURL := urlWithPath(*url, "/oauth/token")
	c := &passwordcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Username:     username,
		Password:     password,
		Endpoint: oauth2.Endpoint{
			TokenURL: tokenURL.String(),
		},
	}
	client := &http.Client{Transport: http.DefaultTransport}
	return &API{
		UnauthenticatedClient: client,
		AuthenticatedClient:   c.Client(context.WithValue(context.Background(), oauth2.HTTPClient, client)),
		TargetURL:             url,
		ZoneID:                zoneID,
	}, nil
}

// NewWithAuthorizationCode builds an API that uses the authorization code
// grant to get a token for use with the UAA API.
//
// You can supply an http.Client because this function has side-effects (a
// token is requested from the target).
//
// If you do not supply an http.Client,
//  http.Client{Transport: http.DefaultTransport}
// will be used.
func NewWithAuthorizationCode(client *http.Client, target string, zoneID string, clientID string, clientSecret string, code string, skipSSLValidation bool) (*API, error) {
	url, err := BuildTargetURL(target)
	if err != nil {
		return nil, err
	}

	tokenURL := urlWithPath(*url, "/oauth/token")
	c := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: tokenURL.String(),
		},
	}

	if client == nil {
		client = &http.Client{Transport: http.DefaultTransport}
	}

	a := &API{
		UnauthenticatedClient: client,
		TargetURL:             url,
		SkipSSLValidation:     skipSSLValidation,
		ZoneID:                zoneID,
	}
	a.ensureTransport(a.UnauthenticatedClient)
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, a.UnauthenticatedClient)
	t, err := c.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	a.AuthenticatedClient = c.Client(ctx, t)

	return a, nil
}