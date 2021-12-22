package xero

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

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
