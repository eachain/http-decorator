package decorator

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

// WithRecovery 防止写逻辑手抖写挂了
func WithRecovery(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e := recover(); e != nil {
				log.Printf("uri = %v, panic error: %v, stack: %s", r.URL.Path, e, debug.Stack())
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
			}
		}()
		handler.ServeHTTP(w, r)
	})
}

// WithHost 只允许通过request.header.host == host的请求
func WithHost(host string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != host {
			log.Printf("uri = %v, bad request: %v",
				r.URL.Path, r.Host)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(http.StatusText(http.StatusBadRequest)))
			return
		}
		handler.ServeHTTP(w, r)
	})
}

// WithMethod 只允许以某一种调用方式调用接口
func WithMethod(method string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			log.Printf("uri = %v, method not allowed: %v",
				r.URL.Path, r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(http.StatusText(http.StatusMethodNotAllowed)))
			return
		}
		handler.ServeHTTP(w, r)
	})
}

// drainBody 复制一份http.Request.Body
func drainBody(b io.ReadCloser) (r1, r2 io.ReadCloser, err error) {
	if b == http.NoBody {
		return http.NoBody, http.NoBody, nil
	}
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(b); err != nil {
		return nil, b, err
	}
	if err = b.Close(); err != nil {
		return nil, b, err
	}
	return ioutil.NopCloser(&buf), ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
}

// responseWriter 复制一份render结果，用于打日志
type responseWriter struct {
	status int
	buf    bytes.Buffer

	http.ResponseWriter // 保证实现http.ResponseWriter接口
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseWriter) Write(p []byte) (int, error) {
	w.buf.Write(p)
	return w.ResponseWriter.Write(p)
}

// WithLogger 在调用接口后打印详细日志
func WithLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		resp := &responseWriter{ResponseWriter: w}
		body := []byte{}

		r1, r2, err := drainBody(r.Body)
		if err != nil {
			log.Printf("uri = %v, read body error: %v", r.URL.Path, err)
			resp.WriteHeader(http.StatusServiceUnavailable)
			resp.Write([]byte(http.StatusText(http.StatusServiceUnavailable)))
		} else {
			r.Body = r1
			body, _ = ioutil.ReadAll(r2)
			handler.ServeHTTP(resp, r)
		}

		if resp.status == 0 {
			resp.status = http.StatusOK
		}

		spent := time.Now().Sub(start)
		log.Printf("uri = %v, method = %v, ip = %v, spent = %v, query = %v, body = %s, status = %v, render = %s",
			r.URL.Path, r.Method, r.RemoteAddr, spent, r.URL.RawQuery, body, resp.status, resp.buf.Bytes())
	})
}
