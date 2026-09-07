package request

import (
	"crypto/tls"
	"reflect"
	"testing"

	"github.com/valyala/fasthttp"
)

func Test_defaultHttpClientOptions(t *testing.T) {
	tests := []struct {
		name string
		want HttpClientOption
	}{
		{
			name: "defaultHttpClientOptions",
			want: HttpClientOption{
				Url:           "",
				HttpMethod:    fasthttp.MethodGet,
				Payload:       make(map[string]interface{}, 0),
				Headers:       make(map[string][]string, 0),
				SkipVerify:    false,
				MinTlsVersion: tls.VersionTLS12,
				Timeout:       30,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := defaultHttpClientOptions(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("defaultHttpClientOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewHttpClient(t *testing.T) {
	type args struct {
		opts []HttpClientOptionFunc
	}

	optsGet := []HttpClientOptionFunc{
		WithUrl("http://localhost:8080"),
		WithHttpMethod(fasthttp.MethodGet),
		WithPayload(make(map[string]interface{})),
		WithHeaders(make(map[string][]string)),
		WithSkipVerify(false),
		WithTimeout(30),
	}
	httpRequestGet := NewHttpClient(optsGet...)

	optsPost := []HttpClientOptionFunc{
		WithUrl("http://localhost:8080"),
		WithHttpMethod(fasthttp.MethodPost),
		WithPayload(make(map[string]interface{})),
		WithHeaders(make(map[string][]string)),
		WithSkipVerify(false),
		WithTimeout(30),
	}
	httpRequestPost := NewHttpClient(optsPost...)

	optsPut := []HttpClientOptionFunc{
		WithUrl("http://localhost:8080"),
		WithHttpMethod(fasthttp.MethodPut),
		WithPayload(make(map[string]interface{})),
		WithHeaders(make(map[string][]string)),
		WithSkipVerify(false),
		WithTimeout(30),
	}
	httpRequestPut := NewHttpClient(optsPut...)

	optsPatch := []HttpClientOptionFunc{
		WithUrl("http://localhost:8080"),
		WithHttpMethod(fasthttp.MethodPatch),
		WithPayload(make(map[string]interface{})),
		WithHeaders(make(map[string][]string)),
		WithSkipVerify(false),
		WithTimeout(30),
	}
	httpRequestPatch := NewHttpClient(optsPatch...)

	optsDelete := []HttpClientOptionFunc{
		WithUrl("http://localhost:8080"),
		WithHttpMethod(fasthttp.MethodDelete),
		WithPayload(make(map[string]interface{})),
		WithHeaders(make(map[string][]string)),
		WithSkipVerify(false),
		WithTimeout(30),
	}
	httpRequestDelete := NewHttpClient(optsDelete...)

	tests := []struct {
		name string
		args args
		want *HttpClient
	}{
		{
			name: "NewHttpClientGet",
			args: args{
				opts: optsGet,
			},
			want: httpRequestGet,
		},
		{
			name: "NewHttpClientPost",
			args: args{
				opts: optsPost,
			},
			want: httpRequestPost,
		},
		{
			name: "NewHttpClientPut",
			args: args{
				opts: optsPut,
			},
			want: httpRequestPut,
		},
		{
			name: "NewHttpClientPatch",
			args: args{
				opts: optsPatch,
			},
			want: httpRequestPatch,
		},
		{
			name: "NewHttpClientDelete",
			args: args{
				opts: optsDelete,
			},
			want: httpRequestDelete,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewHttpClient(tt.args.opts...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHttpClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHttpClient_Call(t *testing.T) {
	type fields struct {
		httpRequest *HttpClient
		err         error
	}
	opts := []HttpClientOptionFunc{
		WithUrl("https://www.google.com"),
		WithHttpMethod(fasthttp.MethodGet),
		WithPayload(make(map[string]interface{})),
		WithHeaders(make(map[string][]string)),
		WithSkipVerify(false),
		WithTimeout(30),
	}
	httpRequest := NewHttpClient(opts...)
	tests := []struct {
		name    string
		fields  fields
		want    *HttpClient
		wantErr bool
	}{
		{
			name: "CallGet",
			fields: fields{
				httpRequest: httpRequest,
				err:         nil,
			},
			want:    httpRequest,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.fields.httpRequest
			got, err := r.Call()
			if (err != nil) != tt.wantErr {
				t.Errorf("HttpClient.Call() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HttpClient.Call() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHttpClient_Close(t *testing.T) {
	type fields struct {
		request          *fasthttp.Request
		response         *fasthttp.Response
		err              error
		HttpClientOption HttpClientOption
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Close",
			fields: fields{
				request:          &fasthttp.Request{},
				response:         &fasthttp.Response{},
				err:              nil,
				HttpClientOption: HttpClientOption{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &HttpClient{
				request:          tt.fields.request,
				response:         tt.fields.response,
				err:              tt.fields.err,
				HttpClientOption: tt.fields.HttpClientOption,
			}
			if err := r.Close(); (err != nil) != tt.wantErr {
				t.Errorf("HttpClient.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHttpClient_GetResponse(t *testing.T) {
	type fields struct {
		request          *fasthttp.Request
		response         *fasthttp.Response
		err              error
		HttpClientOption HttpClientOption
	}
	tests := []struct {
		name   string
		fields fields
		want   *fasthttp.Response
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &HttpClient{
				request:          tt.fields.request,
				response:         tt.fields.response,
				err:              tt.fields.err,
				HttpClientOption: tt.fields.HttpClientOption,
			}
			if got := r.GetResponse(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HttpClient.GetResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHttpClient_GetRawBody(t *testing.T) {
	type fields struct {
		request          *fasthttp.Request
		response         *fasthttp.Response
		err              error
		HttpClientOption HttpClientOption
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &HttpClient{
				request:          tt.fields.request,
				response:         tt.fields.response,
				err:              tt.fields.err,
				HttpClientOption: tt.fields.HttpClientOption,
			}
			got, err := r.GetRawBody()
			if (err != nil) != tt.wantErr {
				t.Errorf("HttpClient.GetRawBody() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HttpClient.GetRawBody() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHttpClient_JSON(t *testing.T) {
	type fields struct {
		request          *fasthttp.Request
		response         *fasthttp.Response
		err              error
		HttpClientOption HttpClientOption
	}
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &HttpClient{
				request:          tt.fields.request,
				response:         tt.fields.response,
				err:              tt.fields.err,
				HttpClientOption: tt.fields.HttpClientOption,
			}
			if err := r.JSON(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("HttpClient.JSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHttpClient_StatusCode(t *testing.T) {
	type fields struct {
		request          *fasthttp.Request
		response         *fasthttp.Response
		err              error
		HttpClientOption HttpClientOption
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &HttpClient{
				request:          tt.fields.request,
				response:         tt.fields.response,
				err:              tt.fields.err,
				HttpClientOption: tt.fields.HttpClientOption,
			}
			if got := r.StatusCode(); got != tt.want {
				t.Errorf("HttpClient.StatusCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHttpClient_Error(t *testing.T) {
	type fields struct {
		request          *fasthttp.Request
		response         *fasthttp.Response
		err              error
		HttpClientOption HttpClientOption
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &HttpClient{
				request:          tt.fields.request,
				response:         tt.fields.response,
				err:              tt.fields.err,
				HttpClientOption: tt.fields.HttpClientOption,
			}
			if err := r.Error(); (err != nil) != tt.wantErr {
				t.Errorf("HttpClient.Error() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWithUrl(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want HttpClientOptionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithUrl(tt.args.url); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithHttpMethod(t *testing.T) {
	type args struct {
		method string
	}
	tests := []struct {
		name string
		args args
		want HttpClientOptionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithHttpMethod(tt.args.method); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithHttpMethod() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithPayload(t *testing.T) {
	type args struct {
		payload map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want HttpClientOptionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithPayload(tt.args.payload); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithPayload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithHeaders(t *testing.T) {
	type args struct {
		headers map[string][]string
	}
	tests := []struct {
		name string
		args args
		want HttpClientOptionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithHeaders(tt.args.headers); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithHeaders() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithSkipVerify(t *testing.T) {
	type args struct {
		skipVerify bool
	}
	tests := []struct {
		name string
		args args
		want HttpClientOptionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithSkipVerify(tt.args.skipVerify); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithSkipVerify() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithMinTlsVersion(t *testing.T) {
	type args struct {
		minTlsVersion uint16
	}
	tests := []struct {
		name string
		args args
		want HttpClientOptionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithMinTlsVersion(tt.args.minTlsVersion); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithMinTlsVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithTimeout(t *testing.T) {
	type args struct {
		timeout int
	}
	tests := []struct {
		name string
		args args
		want HttpClientOptionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithTimeout(tt.args.timeout); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithTimeout() = %v, want %v", got, tt.want)
			}
		})
	}
}
