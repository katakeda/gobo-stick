package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// OpenID ...
type OpenID struct {
	config map[string]interface{}
}

func (openid *OpenID) init() error {
	// Fetch Google's openid configuration JSON
	response, err := http.Get("https://accounts.google.com/.well-known/openid-configuration")
	if err != nil {
		return err
	}
	defer response.Body.Close()
	data, _ := ioutil.ReadAll(response.Body)
	json.Unmarshal(data, &openid.config)

	return nil
}

func (openid *OpenID) get(key string) string {
	endpoint := fmt.Sprintf("%v", openid.config[key])
	return endpoint
}

var openid OpenID
