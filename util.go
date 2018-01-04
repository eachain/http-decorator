package decorator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func parseParam(param interface{}, r *http.Request) error {
	if r.Method == http.MethodGet ||
		strings.Contains(r.Header.Get("Content-Type"), "form") {
		return parseForm(param, r)
	}
	return parseJson(param, r)
}

func parseJson(param interface{}, r *http.Request) error {
	return json.NewDecoder(r.Body).Decode(param)
}

func parseForm(param interface{}, r *http.Request) error {
	v := reflect.Indirect(reflect.ValueOf(param))
	t := v.Type()
	n := v.NumField()
	for i := 0; i < n; i++ {
		f := v.Field(i)
		ft := t.Field(i)
		tag := ft.Tag.Get("form")
		if f.Type().String() == "time.Time" {
			x, err := strconv.ParseInt(r.FormValue(tag), 10, 64)
			if err != nil {
				return fmt.Errorf("field '%v' should be a unix timestamp", tag)
			}
			f.Set(reflect.ValueOf(time.Unix(x, 0)))
			continue
		}
		switch f.Kind() {
		case reflect.Int, reflect.Int64:
			x, err := strconv.ParseInt(r.FormValue(tag), 10, 64)
			if err != nil {
				return fmt.Errorf("field '%v' should be an integer", tag)
			}
			f.SetInt(x)
		case reflect.Float64:
			x, err := strconv.ParseFloat(r.FormValue(tag), 64)
			if err != nil {
				return fmt.Errorf("field '%v' should be a float number", tag)
			}
			f.SetFloat(x)
		case reflect.String:
			f.SetString(r.FormValue(tag))
		case reflect.Slice: // []string
			if f.Type().Elem().Kind() == reflect.String {
				val := r.FormValue(tag)
				var vs []string
				if val != "" {
					// ATTENTION: len(strings.Split("", ",")) == 1
					vs = strings.Split(val, ",")
				}
				f.Set(reflect.ValueOf(vs))
			} else {
				return fmt.Errorf("unsupported field '%v' slice type: %v", tag, f.Type().Elem().Name())
			}
		default:
			return fmt.Errorf("unsupported field '%v' type: %v", tag, f.Type().String())
		}
	}
	return nil
}

type response struct {
	Errno int         `json:"errno"`
	Msg   string      `json:"errmsg"`
	Data  interface{} `json:"data,omitempty"`
}

func render(w http.ResponseWriter, code int, msg string, data interface{}) {
	json.NewEncoder(w).Encode(&response{
		Errno: code,
		Msg:   msg,
		Data:  data,
	})
}

func renderOK(w http.ResponseWriter, data interface{}) {
	render(w, 0, "", data)
}

func renderErr(w http.ResponseWriter, code int, msg string) {
	render(w, code, msg, nil)
}
