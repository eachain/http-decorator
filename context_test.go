package decorator_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/eachain/http-decorator"
)

var (
	ctxMux    *http.ServeMux
	ctxServer *httptest.Server
	ctxClient *http.Client
)

func init() {
	ctxMux = http.NewServeMux()
	ctxServer = httptest.NewServer(ctxMux)
	ctxClient = &http.Client{}
}

func TestWithContext(t *testing.T) {
	const URL = "/context"

	ctxMux.Handle(URL, decorator.WithContext(decorator.HandlerFunc(func(*decorator.Context) {})))

	resp, err := ctxClient.Get(ctxServer.URL + URL)
	if err != nil {
		t.Errorf("test %v error: %v", URL, err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("test %v failed, status code: %v, status: %v", URL, resp.StatusCode, resp.Status)
	}
}

func TestContextParse(t *testing.T) {
	type ReqParam struct {
		A int       `form:"a"`
		B float64   `form:"b"`
		C string    `form:"c"`
		D time.Time `form:"d"`
		E []string  `form:"e"`
	}
	form := url.Values{}
	form.Set("a", "123")
	form.Set("b", "4.56")
	form.Set("c", "xyz")
	form.Set("d", "1512979871")
	form.Set("e", "sn1,sn2,sn3,sn4")
	req, _ := http.NewRequest(http.MethodPost, "http://example.com", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Form = nil

	var param ReqParam
	decorator.WithContext(decorator.HandlerFunc(func(c *decorator.Context) {
		if err := c.Parse(&param); err != nil {
			t.Errorf("parse param error: %v", err)
		}
	})).ServeHTTP(nil, req)

	t.Logf("parse param result: %#v", param)

	if param.A != 123 {
		t.Errorf("param.A = %v", param.A)
	}
	if param.B != 4.56 {
		t.Errorf("param.B = %v", param.B)
	}
	if param.C != "xyz" {
		t.Errorf("param.C = %v", param.C)
	}
	if param.D.Unix() != 1512979871 {
		t.Errorf("param.D = %v", param.D)
	}
	if len(param.E) != 4 {
		t.Errorf("param.E = %v", param.E)
	}
}
