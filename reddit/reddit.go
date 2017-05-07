/*
Package reddit is a client library for the Reddit API: https://www.reddit.com/dev/api/

Only a subset of the API's are implemented
*/
package reddit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

type Reddit struct {
	Client       *http.Client
	UserName     string
	Password     string
	ClientId     string
	ClientSecret string
	Token        Token
	UserAgent    string
}

// Initializesa a new Reddit client for "script" type apps that uses the supplied
// developer username and password, and the app's client id and secret to make
// API calls
func New(username string, password string, clientId string, clientSecret string, userAgent string) (*Reddit, error) {
	tr := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 5 * time.Second,
	}
	client := &http.Client{Transport: tr}

	r := &Reddit{
		Client:       client,
		UserName:     username,
		Password:     password,
		ClientId:     clientId,
		ClientSecret: clientSecret,
		UserAgent:    userAgent,
	}

	err := r.SetAccessToken()
	if err != nil {
		return nil, err
	}
	go r.RefreshAccessToken()

	return r, nil
}

func (r *Reddit) SetAccessToken() (err error) {
	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("username", r.UserName)
	data.Set("password", r.Password)

	request, _ := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", bytes.NewBufferString(data.Encode()))
	request.SetBasicAuth(r.ClientId, r.ClientSecret)
	request.Header.Set("User-Agent", r.UserAgent)

	response, err := r.Client.Do(request)
	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal([]byte(body), &r.Token)
	if err != nil {
		return
	}
	return nil
}

func (r *Reddit) RefreshAccessToken() {
	for {
		time.Sleep(45 * time.Minute)

		err := r.SetAccessToken()
		if err != nil {
			log.Printf("reddit: Could not refresh token: %s", err.Error())
		} else {
			log.Printf("reddit: Successfully refreshed reddit access token")
		}
	}
}

func (r *Reddit) TopListings(subreddit string, time string, limit int) (*ListingsResponse, error) {
	url := fmt.Sprintf("https://oauth.reddit.com/r/%s/top?t=%s&limit=%d", subreddit, time, limit)

	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Set("Authorization", fmt.Sprintf("%s %s", r.Token.TokenType, r.Token.AccessToken))
	request.Header.Set("User-Agent", r.UserAgent)

	response, err := r.Client.Do(request)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var result = new(ListingsResponse)
	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

type Listing struct {
	Url       string `json:"url"`
	Title     string `json:"title"`
	Subreddit string `json:"subreddit"`
}

type ListingsResponse struct {
	ResponseData struct {
		Listings []struct {
			Listing Listing `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

type Token struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}
