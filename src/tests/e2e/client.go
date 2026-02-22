package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func PostJSON(client *http.Client, targetURL string, body interface{}) (*http.Response, error) {
	enc, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, targetURL, bytes.NewReader(enc))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return client.Do(req)
}

func PostJSONWithAuth(client *http.Client, targetURL, bearerToken string, body interface{}) (*http.Response, error) {
	enc, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, targetURL, bytes.NewReader(enc))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	return client.Do(req)
}

func GetWithAuth(client *http.Client, targetURL, bearerToken string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	return client.Do(req)
}

func PutJSONWithAuth(client *http.Client, targetURL, bearerToken string, body interface{}) (*http.Response, error) {
	enc, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPut, targetURL, bytes.NewReader(enc))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	return client.Do(req)
}
