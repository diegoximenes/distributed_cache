package http

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type HTTPError struct {
	StatusCode int
}

func (httpError *HTTPError) Error() string {
	return fmt.Sprintf("%v", httpError.StatusCode)
}

// DoRequest does a request and returns its response.
// Considers that any status code different than 2xx is an error
func DoRequest(request *http.Request) ([]byte, error) {
	httpClient := http.Client{}
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if (response.StatusCode < 200) || (response.StatusCode >= 300) {
		return nil, &HTTPError{
			StatusCode: response.StatusCode,
		}
	}

	return responseBody, nil
}
