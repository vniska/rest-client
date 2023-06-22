package restclient

import (
	"net"
	"net/http"
	"net/http/httptest"
	"crypto/tls"
	"context"
	"io/ioutil"
	"testing"
)

// API returns succeed false and an error message
func TestCallFail(test *testing.T) {
	handler := http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		reqtype := req.Header.Get("Content-Type")
		if reqtype != "application/json" {
			test.Errorf("invalid Content-Type header: %s", reqtype)
		}

		reqmd5 := req.Header.Get("Content-MD5")
		if reqmd5 != "cdefd9b4ca40e984f3482ed3c7ae077a" {
			test.Errorf("invalid MD5 header: %s", reqmd5)
		}

		reqdate := req.Header.Get("Date")
		if reqdate != "2019-02-23T10:03:00+02:00" {
			test.Errorf("invalid Date header: %s", reqdate)
		}

		reqauth := req.Header.Get("Authorization")
		if reqauth != "REALM 123:8923b1bde063c155f8f473b59ea77d2e3134793f9fdef712f1c24f3de6e836ea" {
			test.Errorf("invalid Authorization header: %s", reqauth)
		}

		payload, _ := ioutil.ReadAll(req.Body)
		if string(payload) != "[\"var1\",\"var2\"]" {
			test.Errorf("invalid request payload: %s", string(payload))
		}

		resp.Write([]byte("{\"succeed\":false,\"message\":\"unit test fail\"}"))
	})

	httpclient, srvteardown := createDummyServer(handler);

	defer srvteardown()

	config := Conf {
		123,
		"apisecret",
		"https://api.local",
		1,
		"REALM",
	}

	apiclient, err := NewRestClient(config)

	if err != nil {
		test.Errorf("failed to contruct a restclient: %s", err.Error())
	}

	apiclient.httpclient = httpclient
	apiclient.getTime = func() string {
		return "2019-02-23T10:03:00+02:00"
	}

	data, err := apiclient.Call("unit/test", []string{"var1", "var2"})

	if data != nil {
		test.Error("request was supposed to fail")
	}

	if err == nil {
		test.Error("request did not produce an error")
	}

	if err.Error() != "/api/v1/unit/test: unit test fail" {
		test.Errorf("request did not yield proper error message: %s", err.Error())
	}
}

// API returns succeed true and array of values
func TestCallSuccess1(test *testing.T) {
	handler := http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		reqtype := req.Header.Get("Content-Type")
		if reqtype != "application/json" {
			test.Errorf("invalid Content-Type header: %s", reqtype)
		}

		reqmd5 := req.Header.Get("Content-MD5")
		if reqmd5 != "0a9b61dec51f0560d8bd2a4740dbfe4e" {
			test.Errorf("invalid MD5 header: %s", reqmd5)
		}

		reqdate := req.Header.Get("Date")
		if reqdate != "2019-02-23T11:03:00+02:00" {
			test.Errorf("invalid Date header: %s", reqdate)
		}

		reqauth := req.Header.Get("Authorization")
		if reqauth != "REALM 1234:1c3f8869b2913005048043fa576effd3c003d48b6b6bd72e205b6dedd93c939d" {
			test.Errorf("invalid Authorization header: %s", reqauth)
		}

		payload, _ := ioutil.ReadAll(req.Body)
		if string(payload) != "[\"var3\"]" {
			test.Errorf("invalid request payload: %s", string(payload))
		}

		resp.Write([]byte("{\"succeed\":true,\"result\":[\"val1\",\"val2\"]}"))
	})

	httpclient, srvteardown := createDummyServer(handler);

	defer srvteardown()

	cfg := Conf {
		1234,
		"apisecret2",
		"https://api2.local",
		1,
		"REALM",
	}

	apiclient, err := NewRestClient(cfg)

	if err != nil {
		test.Errorf("failed to contruct a restclient: %s", err.Error())
	}

	apiclient.httpclient = httpclient
	apiclient.getTime = func() string {
		return "2019-02-23T11:03:00+02:00"
	}

	data, err := apiclient.Call("unit/test2", []string{"var3"})

	values, _ := data.([]interface{})
	value1, _ := values[0].(string)
	value2, _ := values[1].(string)

	if value1 != "val1" {
		test.Error("request yielded an unexpected response #1")
	}

	if value2 != "val2" {
		test.Error("request yielded an unexpected response #2")
	}

	if err != nil {
		test.Errorf("request produced an error: %s", err.Error())
	}
}

// API request with multiple params and API returns an array of objects
func TestCallSuccess2(test *testing.T) {
	handler := http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		reqtype := req.Header.Get("Content-Type")
		if reqtype != "application/json" {
			test.Errorf("invalid Content-Type header: %s", reqtype)
		}

		reqmd5 := req.Header.Get("Content-MD5")
		if reqmd5 != "2f7da26fc0796322186a72244f8b8eb4" {
			test.Errorf("invalid MD5 header: %s", reqmd5)
		}

		reqdate := req.Header.Get("Date")
		if reqdate != "2019-02-23T11:03:03+02:00" {
			test.Errorf("invalid Date header: %s", reqdate)
		}

		reqauth := req.Header.Get("Authorization")
		if reqauth != "REALM 1234:43f07efa30cd4ee044f580f6c60cfeafca8fd36211072f77c43512c22289d89f" {
			test.Errorf("invalid Authorization header: %s", reqauth)
		}

		payload, _ := ioutil.ReadAll(req.Body)
		if string(payload) != "[\"var4\",\"var5\"]" {
			test.Errorf("invalid request payload: %s", string(payload))
		}

		resp.Write([]byte("{\"succeed\":true,\"result\":[{\"key1\":\"val1\"},{\"key2\":\"val2\"}]}"))
	})

	httpclient, srvteardown := createDummyServer(handler);

	defer srvteardown()

	cfg := Conf {
		1234,
		"apisecret3",
		"https://api3.local",
		1,
		"REALM",
	}

	apiclient, err := NewRestClient(cfg)

	if err != nil {
		test.Errorf("failed to contruct a restclient: %s", err.Error())
	}

	apiclient.httpclient = httpclient
	apiclient.getTime = func() string {
		return "2019-02-23T11:03:03+02:00"
	}

	data, err := apiclient.Call("unit/test4", []string{"var4", "var5"})

	object_arr, _ := data.([]interface{})
	object1, _ := object_arr[0].(map[string]interface{})
	object2, _ := object_arr[1].(map[string]interface{})

	if object1["key1"].(string) != "val1" {
		test.Error("request yielded an unexpected response #1")
	}

	if object2["key2"].(string) != "val2" {
		test.Error("request yielded an unexpected response #2")
	}

	if err != nil {
		test.Errorf("request produced an error: %s", err.Error())
	}
}

// Creates a dummy server instance which serves mockup HTTP responses
func createDummyServer(handler http.Handler) (*http.Client, func()) {
	server := httptest.NewTLSServer(handler)

	dummyclient := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, network, _ string) (net.Conn, error) {
				return net.Dial(network, server.Listener.Addr().String())
			},
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	return dummyclient, server.Close
}
