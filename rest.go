package rest

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

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

func (r *Rest) Options(path string) *Req {
	return New(r, "OPTIONS", path)
}

func (r *Rest) Get(path string) *Req {
	return New(r, "GET", path)
}

func (r *Rest) Head(path string) *Req {
	return New(r, "HEAD", path)
}

func (r *Rest) Post(path string) *Req {
	return New(r, "POST", path)
}

func (r *Rest) Put(path string) *Req {
	return New(r, "PUT", path)
}

func (r *Rest) Delete(path string) *Req {
	return New(r, "DELETE", path)
}

func (r *Rest) Trace(path string) *Req {
	return New(r, "TRACE", path)
}

func (r *Rest) Connect(path string) *Req {
	return New(r, "CONNECT", path)
}

type Req struct {
	rest    *Rest
	request *http.Request
	path    string
	query   url.Values
	param   url.Values
	json    interface{}
}

func New(rest *Rest, method, path string) *Req {
	request, err := http.NewRequest(method, "/", nil)
	if err != nil {
		panic(err)
	}

	return &Req{
		rest:    rest,
		request: request,
		path:    path,
		query:   make(url.Values),
		param:   make(url.Values),
	}
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

func (req *Req) Json(struc interface{}) *Req {
	req.json = struc
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
		} else if req.json != nil {
			req.ContentType("application/json")
			b, err := json.Marshal(req.json)
			if err != nil {
				return nil, err
			}
			req.Body(bytes.NewReader(b))
		}
	}

	if req.rest.before != nil {
		req.rest.before(req)
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

	res := &Res{response, body}

	if req.rest.after != nil {
		req.rest.after(res)
	}

	return res, nil
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

func (res *Res) Json(struc interface{}) error {
	return json.Unmarshal(res.body, struc)
}
