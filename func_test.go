package decorator_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eachain/http-decorator"
)

var (
	fnMux    *http.ServeMux
	fnServer *httptest.Server
	fnClient *http.Client
)

func init() {
	fnMux = http.NewServeMux()
	fnServer = httptest.NewServer(fnMux)
	fnClient = &http.Client{}
}

func TestWithFunc(t *testing.T) {
	const URL = "/func"

	type Param struct {
		Who string `form:"who"`
	}
	type Resp struct {
		Greeting string `json:"greeting"`
	}

	fnMux.Handle(URL, decorator.WithLogger(decorator.WithFunc(func(req *Param) (*Resp, error) {
		return &Resp{Greeting: "Hello, " + req.Who}, nil
	})))

	resp, err := fnClient.Get(fnServer.URL + URL + "?who=eachain")
	if err != nil {
		t.Errorf("test %v error: %v", URL, err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("test %v failed, status code: %v, status: %v", URL, resp.StatusCode, resp.Status)
	}
}

func TestWithFuncErr(t *testing.T) {
	const URL = "/func/err"

	type Param struct{}
	type Resp struct{}

	const errmsg = "test error"

	fnMux.Handle(URL, decorator.WithFunc(func(req *Param) (*Resp, error) {
		return nil, errors.New(errmsg)
	}))

	resp, err := fnClient.Get(fnServer.URL + URL)
	if err != nil {
		t.Errorf("test %v error: %v", URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("test %v failed, status code: %v, status: %v", URL, resp.StatusCode, resp.Status)
	}

	var buf bytes.Buffer
	var m map[string]interface{}
	err = json.NewDecoder(io.TeeReader(resp.Body, &buf)).Decode(&m)
	if err != nil {
		t.Errorf("test %v failed, json decode error: %v | render: %s", URL, err, buf.Bytes())
	}
	if msg := m["errmsg"].(string); msg != errmsg {
		t.Errorf("test %v failed, errmsg: %v | render: %s", URL, msg, buf.Bytes())
	}
}

func TestWithFuncInvalid(t *testing.T) {
	const URL = "/func/invalid"

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("test %v failed, not panic", URL)
		}
	}()

	type Param struct{}
	fnMux.Handle(URL, decorator.WithFunc(func(req *Param) error {
		return nil
	}))
}
