package rest

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func NewRest() *Rest {
	tr := &http.Transport{
		MaxIdleConnsPerHost: 5,
	}
	client := &http.Client{Transport: tr}

	return &Rest{client: client}
}

type Rest struct {
	before    func(*Req)
	after     func(*Res)
	baseUrl   string
	userAgent string
	client    *http.Client
}

func (r *Rest) BaseUrl(url string) *Rest {
	r.baseUrl = url
	return r
}

func (r *Rest) UserAgent(value string) *Rest {
	r.userAgent = value
	return r
}

func (r *Rest) Before(fn func(req *Req)) *Rest {
	r.before = fn
	return r
}

func (r *Rest) After(fn func(res *Res)) *Rest {
	r.after = fn
	return r
}

/*
func (r *Rest) Options(path string) *Req {
	request, _ := http.NewRequest("OPTIONS", r.baseUrl+path, nil)
	return &Req{rest: r, request: request}
}

func (r *Rest) Get(path string) *Req {
	request, _ := http.NewRequest("GET", r.baseUrl+path, nil)
	return &Req{rest: r, request: request}
}

func (r *Rest) Head(path string) *Req {
	request, _ := http.NewRequest("HEAD", r.baseUrl+path, nil)
	return &Req{rest: r, request: request}
}

func (r *Rest) Post(path string) *Req {
	request, _ := http.NewRequest("POST", r.baseUrl+path, nil)
	return &Req{rest: r, request: request}
}

func (r *Rest) Put(path string) *Req {
	request, _ := http.NewRequest("PUT", r.baseUrl+path, nil)
	return &Req{rest: r, request: request}
}

func (r *Rest) Delete(path string) *Req {
	request, _ := http.NewRequest("DELETE", r.baseUrl+path, nil)
	return &Req{rest: r, request: request}
}

func (r *Rest) Trace(path string) *Req {
	request, _ := http.NewRequest("TRACE", r.baseUrl+path, nil)
	return &Req{rest: r, request: request}
}
*/

func (r *Rest) Get(path string) *Req {
	request, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		panic(err)
	}

	return &Req{
		rest:    r,
		request: request,
		path:    path,
		query:   make(url.Values),
		param:   make(url.Values),
	}
}

func (r *Rest) Connect(path string) *Req {
	request, err := http.NewRequest("CONNECT", "/", nil)
	if err != nil {
		panic(err)
	}

	return &Req{
		rest:    r,
		request: request,
		path:    path,
		query:   make(url.Values),
		param:   make(url.Values),
	}
}

type Req struct {
	rest    *Rest
	request *http.Request
	path    string
	query   url.Values
	param   url.Values
	json    interface{}
}

func (req *Req) Raw() *http.Request {
	return req.request
}

func (req *Req) BasicAuth(username, password string) *Req {
	req.request.SetBasicAuth(username, password)
	return req
}

func (req *Req) Header(key, value string) *Req {
	req.request.Header.Add(key, value)
	return req
}

func (req *Req) Headers(values map[string]string) *Req {
	for key, value := range values {
		req.request.Header.Add(key, value)
	}
	return req
}

func (req *Req) Query(key, value string) *Req {
	req.query.Add(key, value)
	return req
}

func (req *Req) Querys(values map[string]string) *Req {
	for key, value := range values {
		req.query.Add(key, value)
	}
	return req
}

func (req *Req) QueryStruct(value interface{}) *Req {
	return req
}

func (req *Req) Param(key, value string) *Req {
	req.param.Add(key, value)
	return req
}

func (req *Req) Params(values map[string]string) *Req {
	for key, value := range values {
		req.param.Add(key, value)
	}
	return req
}

func (req *Req) ParamsStruct(value interface{}) *Req {
	return req
}

func (req *Req) Json(json interface{}) *Req {
	req.json = json
	return req
}

func (req *Req) ContentType(value string) *Req {
	req.request.Header.Set("Content-Type", value)
	return req
}

func (req *Req) Body(body io.Reader) *Req {
	rc, ok := body.(io.ReadCloser)
	if !ok && body != nil {
		rc = ioutil.NopCloser(body)
	}

	req.request.Body = rc

	if body != nil {
		switch v := body.(type) {
		case *bytes.Buffer:
			req.request.ContentLength = int64(v.Len())
			buf := v.Bytes()
			req.request.GetBody = func() (io.ReadCloser, error) {
				r := bytes.NewReader(buf)
				return ioutil.NopCloser(r), nil
			}
		case *bytes.Reader:
			req.request.ContentLength = int64(v.Len())
			snapshot := *v
			req.request.GetBody = func() (io.ReadCloser, error) {
				r := snapshot
				return ioutil.NopCloser(&r), nil
			}
		case *strings.Reader:
			req.request.ContentLength = int64(v.Len())
			snapshot := *v
			req.request.GetBody = func() (io.ReadCloser, error) {
				r := snapshot
				return ioutil.NopCloser(&r), nil
			}
		default:
			// This is where we'd set it to -1 (at least
			// if body != NoBody) to mean unknown, but
			// that broke people during the Go 1.8 testing
			// period. People depend on it being 0 I
			// guess. Maybe retry later. See Issue 18117.
		}
		// For client requests, Request.ContentLength of 0
		// means either actually 0, or unknown. The only way
		// to explicitly say that the ContentLength is zero is
		// to set the Body to nil. But turns out too much code
		// depends on NewRequest returning a non-nil Body,
		// so we use a well-known ReadCloser variable instead
		// and have the http package also treat that sentinel
		// variable to mean explicitly zero.
		if req.request.GetBody != nil && req.request.ContentLength == 0 {
			req.request.Body = http.NoBody
			req.request.GetBody = func() (io.ReadCloser, error) { return http.NoBody, nil }
		}
	}

	return req
}

func (req *Req) Send() (*Res, error) {

	if req.rest.userAgent != "" {
		req.request.Header.Set("User-Agent", req.rest.userAgent)
	} else {
		req.request.Header.Set("User-Agent", "Rest/0.1")
	}

	if len(req.query) > 0 {
		req.path += "?" + req.query.Encode()
	}

	u, err := url.Parse(req.rest.baseUrl + req.path) // Just url.Parse (url is shadowed for godoc).
	if err != nil {
		return nil, err
	}

	req.request.URL = u

	if req.request.Method == "POST" || req.request.Method == "PUT" || req.request.Method == "PATCH" {
		if len(req.param) > 0 {
			req.ContentType("application/x-www-form-urlencoded")
			req.Body(strings.NewReader(req.param.Encode()))
		}
	}

	response, err := req.rest.client.Do(req.request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return &Res{response, body}, nil
}

type Res struct {
	response *http.Response
	body     []byte
}

func (res *Res) Raw() *http.Response {
	return res.response
}

func (res *Res) Body() []byte {
	return res.body
}

func (res *Res) Status() string {
	return res.response.Status
}

func (res *Res) StatusCode() int {
	return res.response.StatusCode
}

func (res *Res) Header(key string) string {
	return res.response.Header.Get(key)
}

func (res *Res) Json(json interface{}) error {
	return nil
}
