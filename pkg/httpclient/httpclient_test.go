package httpclient

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

const (
	URL        = "http://www.google.com"
	invalidURL = "ht://www.gogle.com"
)

// RequestType
const (
	get = iota
	post
	getWithContext
	postWithContext
)

type HTTPTestCase struct {
	URL         string
	Body        interface{}
	Context     HTTPContext
	RequestType int
	ErrExpected bool //URL specific
}

func TestPosHTTPRequest(t *testing.T) {
	tests := []HTTPTestCase{
		{
			URL:         URL,
			Body:        map[string]string{"data": "value"},
			Context:     HTTPContext{HeaderMap: map[string]string{"Content-Type": ContentTypeJSON}},
			RequestType: post,
			ErrExpected: true,
		},
		{
			URL:         URL,
			Context:     HTTPContext{HeaderMap: map[string]string{"Content-Type": ContentTypeJSON}},
			Body:        map[string]string{"data": "value"},
			RequestType: postWithContext,
			ErrExpected: true,
		},
		{
			URL:         URL,
			Context:     HTTPContext{HeaderMap: map[string]string{"Content-Type": ContentTypeJSON}, Timeout: DefaultTimeout},
			Body:        map[string]string{"data": "value"},
			RequestType: postWithContext,
			ErrExpected: true,
		},
		{
			URL:         URL,
			Context:     HTTPContext{HeaderMap: map[string]string{"Content-Type": ContentTypeJSON}, Retry: 2},
			Body:        map[string]string{"data": "value"},
			RequestType: postWithContext,
			ErrExpected: true,
		},
		{
			URL:         URL,
			Context:     HTTPContext{HeaderMap: map[string]string{"Content-Type": ContentTypeJSON}, Retry: 2, Timeout: 2},
			Body:        map[string]string{"data": "value"},
			RequestType: postWithContext,
			ErrExpected: true,
		},
		{
			URL:         URL,
			Context:     HTTPContext{HeaderMap: map[string]string{"Content-Type": ContentTypeJSON}, Retry: 2, Timeout: 2, QueryParamMap: map[string]interface{}{"a": "b"}},
			Body:        map[string]string{"data": "value"},
			RequestType: postWithContext,
			ErrExpected: true,
		},
		{
			URL:         URL,
			Context:     HTTPContext{QueryParamMap: map[string]interface{}{"a": "b"}},
			RequestType: get,
			ErrExpected: false,
		},
		{
			URL:         URL,
			Context:     HTTPContext{},
			RequestType: getWithContext,
			ErrExpected: false,
		},
		{
			URL:         URL,
			Context:     HTTPContext{Retry: 1},
			RequestType: getWithContext,
			ErrExpected: false,
		},
		{
			URL:         URL,
			Context:     HTTPContext{QueryParamMap: map[string]interface{}{"a": "b"}},
			RequestType: getWithContext,
			ErrExpected: false,
		},
		{
			URL:         URL,
			Context:     HTTPContext{QueryParamMap: map[string]interface{}{"a": "b"}, Timeout: 5, Retry: 1},
			RequestType: getWithContext,
			ErrExpected: false,
		},
	}

	execute(t, tests)
}

func TestNegHttpRequest(t *testing.T) {
	tests := []HTTPTestCase{
		{
			Body:        map[string]string{"a": "b"},
			URL:         invalidURL,
			RequestType: post,
			ErrExpected: true,
		},
		{
			Body:        make(chan int),
			URL:         URL,
			RequestType: post,
			ErrExpected: true,
		},
		{
			URL:         invalidURL,
			Context:     HTTPContext{HeaderMap: map[string]string{"Content-Type": ContentTypeJSON}},
			Body:        nil,
			ErrExpected: true,
		},
		{
			URL:         invalidURL,
			Context:     HTTPContext{Timeout: 1},
			Body:        map[string]string{"data": "value"},
			ErrExpected: true,
		},
		{
			URL:         "http://invalidcaseurl.com",
			Context:     HTTPContext{HeaderMap: map[string]string{"Content-Type": ContentTypeJSON}, Retry: 1, Timeout: 1},
			Body:        map[string]string{"data": "value"},
			ErrExpected: true,
		},
		{
			URL:         "invalidurl",
			Context:     HTTPContext{HeaderMap: map[string]string{"Content-Type": ContentTypeJSON}, Retry: 1, Timeout: 1},
			Body:        nil,
			ErrExpected: true,
		},
		{
			URL:         invalidURL,
			RequestType: get,
			ErrExpected: true,
		},
		{
			URL:         invalidURL,
			Context:     HTTPContext{Timeout: 1},
			RequestType: getWithContext,
			ErrExpected: true,
		},
		{
			URL:         invalidURL,
			Context:     HTTPContext{QueryParamMap: map[string]interface{}{"a": "b"}, Timeout: 1},
			RequestType: getWithContext,
			ErrExpected: true,
		},
		{
			URL:         "http://invalidcaseurl.com", //Non reachable url
			Context:     HTTPContext{QueryParamMap: map[string]interface{}{}, Timeout: 1, Retry: 1},
			RequestType: getWithContext,
			ErrExpected: true,
		},
		{
			URL:         "",
			Context:     HTTPContext{Timeout: 1, Retry: 1},
			RequestType: getWithContext,
			ErrExpected: true,
		},
	}

	execute(t, tests)

}

func TestFormatURL(t *testing.T) {
	testCases := []struct {
		inputURL, expectedURL string
	}{
		{
			inputURL:    "www.google.com",
			expectedURL: "https://www.google.com",
		},
		{
			inputURL:    "www.google.com?a=b&amp;c=d",
			expectedURL: "https://www.google.com?a=b&c=d",
		},
	}

	for i, tc := range testCases {
		t.Logf("TestFormatURL::TestCase #%v", i)
		fmtURL := formatURL(tc.inputURL)
		assert.Equal(t, fmtURL, tc.expectedURL, "Expected %v, Got %v", tc.expectedURL, fmtURL)
	}
}

func execute(t *testing.T, tests []HTTPTestCase) {
	for i, tc := range tests {
		tc := tc
		t.Run(fmt.Sprintf("#%v", i), func(st *testing.T) {
			st.Parallel()
			var (
				res []byte
				err error
			)
			switch tc.RequestType {
			case get:
				res, err = Get(tc.URL)
			case post:
				res, err = Post(tc.Body, tc.URL, tc.Context.HeaderMap)
			case getWithContext:
				res, err = GetWithContext(tc.URL, tc.Context)
			case postWithContext:
				res, err = PostWithContext(tc.Body, tc.URL, tc.Context)
			}
			if tc.ErrExpected {
				assert.NotNil(t, err, "Error: %+v", err)
			} else {
				assert.Nil(t, err, "Error: %+v", err)
				assert.NotNil(t, res, "Response: %v", string(res))
			}
		})
	}
}
