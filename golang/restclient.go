// Package restclient provides a REST client to communicate with the Liana Technologies REST API.
package restclient

import (
	"fmt"
	"net/http"
	"encoding/json"
	"encoding/hex"
	"strings"
	"crypto/md5"
	"crypto/sha256"
	"crypto/hmac"
	"time"
	"io/ioutil"
	"errors"
)

type Conf struct {
	Userid int
	Secret string
	Apiurl string
	Apiversion int
	Apirealm string
}

type restclient struct {
	cfg Conf

	// Helpers
	getTime func() string // overridable in tests!
	httpclient *http.Client

	// Private vars
	endpoint string
	bodystr string
	hashstr string
	timestr string
}

func NewRestClient(cfg Conf) (restclient, error) {
	// The time string cannot be made static here as we can't know how far into the
	// future the restclient instance lives.
	getTime := func() string {
		return time.Now().Format(time.RFC3339)
	}

	httpclient := &http.Client{
		Timeout: 60 * time.Second,
	}

	return restclient {cfg, getTime, httpclient, "", "", "" ,""}, nil
}

// Call is used to perform a call to the RESTful API.
func (this restclient) Call(path string, params interface{}, inputMethod ...string) (interface{}, error) {
	bodyjson, err := json.Marshal(params)

	if err != nil {
		return nil, err
	}

	this.bodystr = string(bodyjson)

	method := "POST"
	if len(inputMethod) > 0 {
		method = inputMethod[0]
	}
	if method == "GET" {
		this.bodystr = "";
	}

	hash := md5.New()
	hash.Write([]byte(string(this.bodystr)))
	this.hashstr = hex.EncodeToString(hash.Sum(nil))
	this.endpoint = fmt.Sprintf("/api/v%d/%s", this.cfg.Apiversion, path)
	this.timestr = this.getTime();

	req, err := this.createRequest(method)

	if err != nil {
		return nil, err
	}

	resp, err := this.httpclient.Do(req)

	if err != nil {
		return nil, err
	}

	return this.handleResults(resp)
}

// Form the http.Request to fetch the response data
func (this restclient) createRequest(method string) (*http.Request, error) {
	req, err := http.NewRequest(
		method,
		this.cfg.Apiurl + this.endpoint,
		strings.NewReader(this.bodystr),
	)

	if err != nil {
		return nil, err
	}

	sign := hmac.New(sha256.New, []byte(this.cfg.Secret))
	sign.Write([]byte(fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		method,
		this.hashstr,
		"application/json",
		this.timestr,
		this.bodystr,
		this.endpoint,
	)))

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-MD5", this.hashstr)
	req.Header.Add("Date", this.timestr)
	req.Header.Add(
		"Authorization",
		fmt.Sprintf("%s %d:%s", this.cfg.Apirealm, this.cfg.Userid, hex.EncodeToString(sign.Sum(nil))),
	)

	return req, nil
}

// Preprocesses an API response
// Response json is decoded and checked for the success value.
func (this restclient) handleResults(resp *http.Response) (interface{}, error) {
	defer resp.Body.Close()

	var results map[string]interface{}

	bodystr, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	json.Unmarshal(bodystr, &results)

	if this.cfg.Apiversion == 1 || this.cfg.Apiversion == 2 {
		var succeed, ok = results["succeed"]
		if ! ok {
			return nil, errors.New("unexpected response from API: " + string(bodystr))
		}
		if ! succeed.(bool) {
			return nil, errors.New(this.endpoint + ": " + results["message"].(string))
		}
		return results["result"], nil
	} else if this.cfg.Apiversion == 3 {
		var _, ok = results["items"]
		if ! ok {
			return nil, errors.New("unexpected response from API: " + string(bodystr))
		}
		return results["items"], nil
	}

	return nil, errors.New(fmt.Sprintf("unexpected api version %+v", this.cfg.Apiversion))
}

