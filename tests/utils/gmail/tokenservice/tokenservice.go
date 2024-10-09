// Copyright (c) 2024 Proton AG
//
// This file is part of Proton Mail Bridge.Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package tokenservice

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

const (
	NexusTestingPwd  = "NEXUS_TESTING_PWD"
	NexusTestingUser = "NEXUS_TESTING_USER"
)

var nexusAccessVerificationURL = os.Getenv("NEXUS_ACCESS_VERIFICATION_URL")
var nexusCredentialsURL = os.Getenv("NEXUS_GMAIL_CREDENTIALS_URL")
var nexusTokenURL = os.Getenv("NEXUS_GMAIL_TOKEN_URL")

var gmailScopes = []string{gmail.GmailComposeScope, gmail.GmailInsertScope, gmail.GmailLabelsScope, gmail.MailGoogleComScope, gmail.GmailMetadataScope, gmail.GmailModifyScope, gmail.GmailSendScope}

func fetchBytes(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch data from url %v: %v", url, err)
	}
	defer func() { _ = res.Body.Close() }()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body for url %v: %v", url, err)
	}

	return body, nil
}

func fetchConfig() (*oauth2.Config, error) {
	data, err := fetchBytes(nexusCredentialsURL)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch credentials: %v", err)
	}

	config, err := google.ConfigFromJSON(data, gmailScopes...)
	if err != nil {
		return nil, fmt.Errorf("unable to parse credentials to gmail config: %v", err)
	}

	return config, err
}

func fetchToken() (*oauth2.Token, error) {
	data, err := fetchBytes(nexusTokenURL)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch token: %v", err)
	}

	token := &oauth2.Token{}
	if err = json.Unmarshal(data, token); err != nil {
		return nil, fmt.Errorf("error when unmarshaling token: %v", err)
	}

	return token, nil
}

func refreshToken(config *oauth2.Config, token *oauth2.Token) (*oauth2.Token, error) {
	ctx := context.Background()
	tokenSource := config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("error when refreshing access token: %v", err)
	}
	return newToken, nil
}

func getNexusCreds() string {
	credentials := os.Getenv(NexusTestingUser) + ":" + os.Getenv(NexusTestingPwd)
	return base64.StdEncoding.EncodeToString([]byte(credentials))
}

func pushToNexus(url string, data []byte) error {
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("error creating put request to nexus: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	encodedCredentials := getNexusCreds()
	req.Header.Set("Authorization", "Basic "+encodedCredentials)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making put request to nexus: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code when making put request to nexus: %v", resp.StatusCode)
	}

	return nil
}

func uploadTokenToNexus(token *oauth2.Token) error {
	jsonData, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("error when encoding access token to json: %v", err)
	}

	return pushToNexus(nexusTokenURL, jsonData)
}

func verifyNexusAccess() error {
	return pushToNexus(nexusAccessVerificationURL, nil)
}

func checkTokenValidityAndRefresh(config *oauth2.Config, token *oauth2.Token) (*oauth2.Token, error) {
	// Validate token (check if it has expired, or is 5 minutes from expiring) and refresh
	timeTillExpiry := time.Until(token.Expiry)
	if !token.Valid() || timeTillExpiry < 5*time.Minute {
		token, err := refreshToken(config, token)
		if err != nil {
			return nil, err
		}

		if err = uploadTokenToNexus(token); err != nil {
			return nil, fmt.Errorf("unable to upload token to nexus: %v", err)
		}
	}
	return token, nil
}

func LoadGmailClient(ctx context.Context) (*http.Client, error) {
	err := verifyNexusAccess()
	if err != nil {
		log.Fatalf("error occurred when verifying nexus access, check your credentials: %v", err)
	}

	config, err := fetchConfig()
	if err != nil {
		return nil, fmt.Errorf("issue obtaining oauth config: %v", err)
	}

	token, err := fetchToken()
	if err != nil {
		return nil, fmt.Errorf("issue obtaining oauth token: %v", err)
	}

	token, err = checkTokenValidityAndRefresh(config, token)
	if err != nil {
		return nil, fmt.Errorf("error checking token validity: %v", err)
	}

	client := config.Client(ctx, token)
	return client, nil
}
