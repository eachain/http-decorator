package decorator_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/eachain/http-decorator"
)

var (
	decMux    *http.ServeMux
	decServer *httptest.Server
	decClient *http.Client
)

func init() {
	decMux = http.NewServeMux()
	decServer = httptest.NewServer(decMux)
	decClient = &http.Client{}
}

func TestWithRecovery(t *testing.T) {
	const URL = "/recovery"
	decMux.Handle(URL, decorator.WithRecovery(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("test panic")
	})))

	resp, err := decClient.Get(decServer.URL + URL)
	if err != nil {
		t.Errorf("test %v error: %v", URL, err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("test %v failed, status code: %v, status: %v", URL, resp.StatusCode, resp.Status)
	}
}

func TestWithHost(t *testing.T) {
	const URL = "/host"
	decMux.Handle(URL, decorator.WithHost("www.eachain.com", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})))

	testDo := func(host string, status int) {
		req, err := http.NewRequest("GET", decServer.URL+URL, nil)
		if err != nil {
			t.Errorf("test %v new request error: %v", URL, err)
		}
		req.Host = host
		resp, err := decClient.Do(req)
		if err != nil {
			t.Errorf("test %v error: %v", URL, err)
		}
		resp.Body.Close()

		if resp.StatusCode != status {
			t.Errorf("test %v failed, status code: %v, status: %v", URL, resp.StatusCode, resp.Status)
		}
	}

	testDo("www.eachain.com", http.StatusOK)
	testDo("www.yichen.com", http.StatusBadRequest)
	testDo("", http.StatusBadRequest)
}

func TestWithMethod(t *testing.T) {
	const URL = "/method"
	decMux.Handle(URL, decorator.WithMethod("GET", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})))

	testDo := func(do func(rawURL string) (*http.Response, error), status int) {
		resp, err := do(decServer.URL + URL)
		if err != nil {
			t.Errorf("test %v error: %v", URL, err)
		}
		resp.Body.Close()

		if resp.StatusCode != status {
			t.Errorf("test %v failed, status code: %v, status: %v", URL, resp.StatusCode, resp.Status)
		}
	}

	testDo(decClient.Get, http.StatusOK)
	testDo(decClient.Head, http.StatusMethodNotAllowed)
}

func TestWithLogger(t *testing.T) {
	const URL = "/logger"
	const BODY = "Hello world"
	decMux.Handle(URL, decorator.WithLogger(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(BODY))
	})))

	resp, err := decClient.Post(decServer.URL+URL+"?a=123", "", strings.NewReader("request-body"))
	if err != nil {
		t.Errorf("test %v error: %v", URL, err)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("test %v failed, status code: %v, status: %v", URL, resp.StatusCode, resp.Status)
	}

	if string(body) != BODY {
		t.Errorf("test %v failed, body: %s", body)
	}
}
