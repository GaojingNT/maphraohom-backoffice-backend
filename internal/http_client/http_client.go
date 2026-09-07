package request

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"slices"
	"time"

	"github.com/goccy/go-json"
	"github.com/valyala/fasthttp"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
)

type HttpClient struct {
	request  *fasthttp.Request
	response *fasthttp.Response
	err      error
	HttpClientOption
}

type HttpClientOption struct {
	Url           string
	HttpMethod    string
	Payload       map[string]interface{}
	Headers       map[string][]string
	SkipVerify    bool
	MinTlsVersion uint16
	Timeout       int
}

func defaultHttpClientOptions() HttpClientOption {
	return HttpClientOption{
		Url:           "",
		HttpMethod:    fasthttp.MethodGet,
		Payload:       make(map[string]interface{}, 0),
		Headers:       make(map[string][]string, 0),
		SkipVerify:    false,
		MinTlsVersion: tls.VersionTLS12,
		Timeout:       30,
	}
}

type HttpClientOptionFunc func(*HttpClientOption)

func NewHttpClient(opts ...HttpClientOptionFunc) *HttpClient {
	httpRequestOptions := defaultHttpClientOptions()

	for _, fn := range opts {
		fn(&httpRequestOptions)
	}

	return &HttpClient{
		HttpClientOption: httpRequestOptions,
	}
}

func (r *HttpClient) Call() (*HttpClient, error) {
	var err error

	// Create fasthttp client
	client := &fasthttp.Client{
		TLSConfig: &tls.Config{
			MinVersion:         r.MinTlsVersion,
			InsecureSkipVerify: r.SkipVerify,
		},
		MaxConnWaitTimeout: time.Duration(r.Timeout) * time.Second,
	}

	// Make request
	r.request = fasthttp.AcquireRequest()

	// Set request headers
	for key, value := range r.Headers {
		for _, v := range value {
			r.request.Header.Add(key, v)
		}
	}

	// Set request transport for skip tls verification
	if r.SkipVerify {
		r.request.Header.Set("X-Forwarded-Proto", "https")
		r.request.Header.Set("X-Forwarded-Port", "443")
	}

	// Set request method and URL
	r.request.Header.SetMethod(r.HttpMethod)
	r.request.SetRequestURI(r.Url)

	// Set request body
	var payloadBytes []byte
	// Create a buffer
	requestBody := new(bytes.Buffer)
	// Create a new multipart writer
	multipartWriter := multipart.NewWriter(requestBody)

	if slices.Contains([]string{http.MethodPost, http.MethodPut}, r.HttpMethod) {
		if r.Headers["Content-Type"][0] == "application/json" {
			payloadBytes, err = json.Marshal(r.Payload)
			if err != nil {
				r.err = err
			}
			requestBody = bytes.NewBuffer(payloadBytes)
		} else if r.Headers["Content-Type"][0] == "application/x-www-form-urlencoded" {
			// Create a new url.Values
			params := url.Values{}

			// Set request payload for form data
			for key, payloadValue := range r.Payload {
				params.Add(key, fmt.Sprintf("%v", payloadValue))
			}

			// Set request body
			requestBody = bytes.NewBufferString(params.Encode())
		} else if r.Headers["Content-Type"][0] == "multipart/form-data" {
			// Set request payload for form data
			for key, payloadValue := range r.Payload {
				// Switch case by data type
				switch pVal := payloadValue.(type) {
				case string:
					_ = multipartWriter.WriteField(key, pVal)
				case int:
					_ = multipartWriter.WriteField(key, fmt.Sprintf("%d", pVal))
				case int64:
					_ = multipartWriter.WriteField(key, fmt.Sprintf("%d", pVal))
				case float64:
					_ = multipartWriter.WriteField(key, fmt.Sprintf("%f", pVal))
				case bool:
					_ = multipartWriter.WriteField(key, fmt.Sprintf("%t", pVal))
				case *os.File:
					// Add file to form data
					fileWriter, _ := multipartWriter.CreateFormFile(key, pVal.Name())
					_, _ = io.Copy(fileWriter, pVal)
				default:
					continue
				}
			}

			// Close the multipart writer to finalize the form data
			multipartWriter.Close()
		} else {
			// Return error if content type is not supported
			r.err = exception.ErrRequestContentTypeNotSupported
		}
	} else {
		// HTTP GET request

		// Create a new url.Values
		params := url.Values{}

		// Set request payload for query string
		for key, payloadValue := range r.Payload {
			params.Add(key, fmt.Sprintf("%v", payloadValue))
		}

		// Set request url with query string
		r.Url = fmt.Sprintf("%s?%s", r.Url, params.Encode())
	}

	// Set request body
	r.request.SetBody(requestBody.Bytes())

	// Make response
	r.response = fasthttp.AcquireResponse()

	err = client.Do(r.request, r.response)
	if err != nil {
		r.err = err
	}

	return r, err
}

func (r *HttpClient) Close() error {
	fasthttp.ReleaseRequest(r.request)
	fasthttp.ReleaseResponse(r.response)

	return nil
}

func (r *HttpClient) GetResponse() *fasthttp.Response {
	return r.response
}

func (r *HttpClient) GetRawBody() ([]byte, error) {
	// Decompress the response
	contentEncoding := r.response.Header.Peek("Content-Encoding")
	var body []byte
	if bytes.EqualFold(contentEncoding, []byte("gzip")) {
		body, _ = r.response.BodyGunzip()
	} else {
		body = r.response.Body()
	}

	return body, nil
}

func (r *HttpClient) JSON(v interface{}) error {
	// Get raw response body
	body, err := r.GetRawBody()
	if err != nil {
		return err
	}

	// Unmarshal response body to struct
	err = json.Unmarshal(body, v)
	if err != nil {
		return err
	}

	return nil
}

func (r *HttpClient) StatusCode() int {
	return r.response.StatusCode()
}

func (r *HttpClient) Error() error {
	return r.err
}

func WithUrl(url string) HttpClientOptionFunc {
	return func(o *HttpClientOption) {
		o.Url = url
	}
}

func WithHttpMethod(method string) HttpClientOptionFunc {
	return func(o *HttpClientOption) {
		o.HttpMethod = method
	}
}

func WithPayload(payload map[string]interface{}) HttpClientOptionFunc {
	return func(o *HttpClientOption) {
		o.Payload = payload
	}
}

func WithHeaders(headers map[string][]string) HttpClientOptionFunc {
	return func(o *HttpClientOption) {
		o.Headers = headers
	}
}

func WithSkipVerify(skipVerify bool) HttpClientOptionFunc {
	return func(o *HttpClientOption) {
		o.SkipVerify = skipVerify
	}
}

func WithMinTlsVersion(minTlsVersion uint16) HttpClientOptionFunc {
	return func(o *HttpClientOption) {
		o.MinTlsVersion = minTlsVersion
	}
}

func WithTimeout(timeout int) HttpClientOptionFunc {
	return func(o *HttpClientOption) {
		o.Timeout = timeout
	}
}
