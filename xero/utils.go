package xero

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"golang.org/x/oauth2"
)

type XeroClient struct {
	// path where the token will be saved
	TokenPath string
	TenantId  string
	Token     *oauth2.Token
	Client    *http.Client
}

type Connections []struct {
	ID             string `json:"id"`
	AuthEventID    string `json:"authEventId"`
	TenantID       string `json:"tenantId"`
	TenantType     string `json:"tenantType"`
	TenantName     string `json:"tenantName"`
	CreatedDateUtc string `json:"createdDateUtc"`
	UpdatedDateUtc string `json:"updatedDateUtc"`
}

// Store the tenant ID for the named organisation
func (cli *XeroClient) StoreTenantId(tenantName string) error {
	resp, err := cli.Client.Get("https://api.xero.com/connections")
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("error while fetching connections: %v", resp.Status)
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	connections := Connections{}
	err = json.Unmarshal(body, &connections)
	if err != nil {
		return fmt.Errorf("error while decoding connections json: %v", err)
	}

	for i := 0; i < len(connections); i++ {
		// if no tenant name is given, use the first one
		if tenantName == "" || connections[i].TenantName == tenantName {
			cli.TenantId = connections[i].TenantID
			return nil
		}
	}

	return fmt.Errorf("could not find organisation with name: %s", tenantName)
}

// Ensure that the http client has a fresh token
func (cli *XeroClient) EnsureFreshToken() error {
	// has the token expired, or will it expire in the next 10 seconds?
	if cli.Token.Expiry.Before(time.Now().Add(time.Duration(10) * time.Second)) {
		// refreshes the token, and updates it on the client directly
		newToken, err := cli.Client.Transport.(*oauth2.Transport).Source.Token()
		if err != nil {
			return err
		}
		if newToken.AccessToken != cli.Token.AccessToken {
			cli.Token = newToken
			saveOauthToken(newToken, cli.TokenPath)
		}
	}
	return nil
}

// Save oauth token information to a file, for re-use later
func saveOauthToken(token *oauth2.Token, tokenPath string) error {
	file, _ := json.MarshalIndent(token, "", " ")
	return os.WriteFile(tokenPath, file, 0644)
}

// Load oauth token information from file
func loadOauthToken(tokenPath string) (*oauth2.Token, error) {
	file, err := os.ReadFile(tokenPath)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	token := oauth2.Token{}
	err = json.Unmarshal([]byte(file), &token)
	return &token, err
}

func connect(ctx context.Context, d *plugin.QueryData) (*XeroClient, error) {
	// Load connection from cache, which preserves throttling protection etc
	cacheKey := "xero"
	if cachedData, ok := d.ConnectionManager.Cache.Get(cacheKey); ok {
		client := cachedData.(*XeroClient)
		// ensure the token is fresh
		err := client.EnsureFreshToken()
		return client, err
	}

	clientId := os.Getenv("XERO_CLIENT_ID")
	clientSecret := os.Getenv("XERO_CLIENT_SECRET")
	tenantName := os.Getenv("XERO_TENANT_NAME")
	code := os.Getenv("XERO_OAUTH_CODE")
	tokenPath := "~/.steampipe/internal/xero-oauth-token.json"
	redirectURL := ""

	// get config from file
	pluginConfig := GetConfig(d.Connection)
	if pluginConfig.ClientId != nil {
		clientId = *pluginConfig.ClientId
	}
	if pluginConfig.ClientSecret != nil {
		clientSecret = *pluginConfig.ClientSecret
	}
	if pluginConfig.TenantName != nil {
		tenantName = *pluginConfig.TenantName
	}
	if pluginConfig.OauthCode != nil {
		code = *pluginConfig.OauthCode
	}
	if pluginConfig.RedirectURL != nil {
		redirectURL = *pluginConfig.RedirectURL
	}
	if pluginConfig.OauthTokenPath != nil {
		tokenPath = *pluginConfig.OauthTokenPath
	}
	tokenPath = resolvePath(tokenPath)

	if clientId == "" {
		return nil, fmt.Errorf("xero client_id must be specified")
	}
	if clientSecret == "" {
		return nil, fmt.Errorf("xero client_secret must be specified")
	}
	if redirectURL == "" {
		return nil, fmt.Errorf("xero redirect_url must be specified")
	}

	oauthConfig := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"accounting.transactions.read", "offline_access"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://login.xero.com/identity/connect/authorize",
			TokenURL: "https://identity.xero.com/connect/token",
		},
	}

	token, err := loadOauthToken(tokenPath)
	if err != nil {
		return nil, err
	}

	if token == nil {
		url := oauthConfig.AuthCodeURL("steampipe-plugin-xero", oauth2.AccessTypeOffline)
		if code == "" {
			return nil, fmt.Errorf("xero oauth_code must be specified; visit %s to authenticate and get a code", url)
		}

		// do exchange to get a token since there is no existing token
		plugin.Logger(ctx).Info("Getting new oauth token...")
		token, err = oauthConfig.Exchange(ctx, code)
		if err != nil {
			// the oauth code is invalid
			return nil, fmt.Errorf("error getting new oauth token (the oauth_code is probably old): %v. Visit %s to authenticate and get a new code", err, url)
		}
		plugin.Logger(ctx).Info("Received and saving new oauth token.")
		saveOauthToken(token, tokenPath)
	} else {
		plugin.Logger(ctx).Info("Using saved oauth token.")
	}

	client := &XeroClient{
		TokenPath: tokenPath,
		Token:     token,
		Client:    oauthConfig.Client(ctx, token),
	}

	// ensure the oauth tokens are still valid
	err = client.EnsureFreshToken()
	if err != nil {
		// the oauth refresh token is probably out of date, throw it away, so that the next time
		// we try to connect we start from scratch
		os.Remove(tokenPath)
		return client, err
	}

	// lookup and store the tenant id
	err = client.StoreTenantId(tenantName)
	if err != nil {
		return client, err
	}

	// Save to cache
	d.ConnectionManager.Cache.Set(cacheKey, client)

	return client, nil
}

// Expand ~ and ~/ in path
// See https://stackoverflow.com/questions/17609732/expand-tilde-to-home-directory/17617721
func resolvePath(path string) string {
	usr, _ := user.Current()
	dir := usr.HomeDir

	if path == "~" {
		// In case of "~", which won't be caught by the "else if"
		path = dir
	} else if strings.HasPrefix(path, "~/") {
		// Use strings.HasPrefix so we don't match paths like
		// "/something/~/something/"
		path = filepath.Join(dir, path[2:])
	}

	return path
}
