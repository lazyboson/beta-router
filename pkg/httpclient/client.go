package httpclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

//HTTPContext - to specify additional http request attributes
type HTTPContext struct {
	Timeout       uint
	Retry         uint
	QueryParamMap map[string]interface{}
	HeaderMap     map[string]string
}

//Post - post the message to the specified address
func Post(msgBody interface{}, HTTPUrl string, headermap map[string]string) ([]byte, error) {
	return sendHTTPRequest(POST, HTTPUrl, msgBody, HTTPContext{HeaderMap: headermap})
}

//PostWithContext - post the message to the specified address with context
func PostWithContext(msgBody interface{}, HTTPUrl string, ctx HTTPContext) ([]byte, error) {
	return sendHTTPRequest(POST, HTTPUrl, msgBody, ctx)
}

//Get - send Get request to the specified address
func Get(HTTPUrl string) ([]byte, error) {
	return sendHTTPRequest(GET, HTTPUrl, nil, HTTPContext{})
}

//GetWithContext - send Get request to the specified address with context
func GetWithContext(HTTPUrl string, ctx HTTPContext) ([]byte, error) {
	return sendHTTPRequest(GET, HTTPUrl, nil, ctx)
}

//Delete - send Delete request to the specified address
func Delete(HTTPUrl string) ([]byte, error) {
	return sendHTTPRequest(DELETE, HTTPUrl, nil, HTTPContext{})
}

func sendHTTPRequest(HTTPMethod, HTTPUrl string, requestBody interface{}, ctx HTTPContext) ([]byte, error) {
	//change default timeout
	if ctx.Timeout == 0 {
		ctx.Timeout = DefaultTimeout
	}
	ctx.Retry++ // total http(s) requests will be 1 + retryCount
	client := &http.Client{
		Timeout: time.Second * time.Duration(ctx.Timeout),
	}

	var (
		req *http.Request
		err error
	)

	HTTPUrl = formatURL(HTTPUrl)
	switch HTTPMethod {
	case GET:
		//change query param map
		if ctx.QueryParamMap != nil && len(ctx.QueryParamMap) > 0 {
			params := url.Values{}
			for paramName, paramValue := range ctx.QueryParamMap {
				params.Add(paramName, fmt.Sprintf("%v", paramValue))
			}
			HTTPUrl += "?" + params.Encode()
		}
		req, err = http.NewRequest(GET, HTTPUrl, nil)
	case POST:
		buf := new(bytes.Buffer)
		err = json.NewEncoder(buf).Encode(requestBody)
		if err != nil {
			return nil, err.(*json.UnsupportedTypeError)
		}
		req, err = http.NewRequest(POST, HTTPUrl, buf)
	case DELETE:
		req, err = http.NewRequest(DELETE, HTTPUrl, nil)
	default:
		return nil, errors.New("http method not supported")
	}
	if err != nil {
		return nil, err
	}

	if ctx.HeaderMap != nil {
		for key, element := range ctx.HeaderMap {
			req.Header.Set(key, element)
		}
	}

	var resp *http.Response
	for ctx.Retry > 0 {
		ctx.Retry--

		resp, err = client.Do(req)
		if err != nil {
			continue //TODO: can retry only for 5XX errors
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		//close body when finished
		resp.Body.Close()

		if resp.StatusCode < http.StatusMultipleChoices {
			return body, nil
		} else {
			return nil, errors.New(string(body))
		}

	}

	return nil, errors.New(maxRetryErrMsg)
}
