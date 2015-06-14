package rainmaker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/pivotal-cf-experimental/rainmaker/internal/network"
)

type Config struct {
	SkipVerifySSL bool
	Host          string
}

type Client struct {
	Config           Config
	Organizations    *OrganizationsService
	Spaces           *SpacesService
	Users            *UsersService
	ServiceInstances *ServiceInstancesService
}

func NewClient(config Config) Client {
	return Client{
		Config:           config,
		Organizations:    NewOrganizationsService(config),
		Spaces:           NewSpacesService(config),
		Users:            NewUsersService(config),
		ServiceInstances: NewServiceInstancesService(config),
	}
}

type requestArguments struct {
	Method                string
	Path                  string
	Query                 url.Values
	Token                 string
	Body                  interface{}
	AcceptableStatusCodes []int
}

func (client Client) makeRequest(requestArgs requestArguments) (int, []byte, error) {
	if requestArgs.AcceptableStatusCodes == nil {
		panic("No acceptable status codes were assigned in the request arguments")
	}

	jsonBody, err := json.Marshal(requestArgs.Body)
	if err != nil {
		return 0, []byte{}, newRequestBodyMarshalError(err)
	}

	requestURL, err := url.Parse(client.Config.Host)
	if err != nil {
		return 0, []byte{}, newRequestConfigurationError(err)
	}

	requestURL.Path = requestArgs.Path
	requestURL.RawQuery = requestArgs.Query.Encode()

	request, err := http.NewRequest(requestArgs.Method, requestURL.String(), bytes.NewBuffer(jsonBody))
	if err != nil {
		return 0, []byte{}, newRequestConfigurationError(err)
	}

	request.Header.Set("Authorization", "Bearer "+requestArgs.Token)

	client.printRequest(request)

	networkClient := network.GetClient(network.Config{SkipVerifySSL: client.Config.SkipVerifySSL})
	response, err := networkClient.Do(request)
	if err != nil {
		return 0, []byte{}, newRequestHTTPError(err)
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, []byte{}, newResponseReadError(err)
	}

	client.printResponse(response.StatusCode, responseBody)

	if response.StatusCode == 404 {
		return 0, []byte{}, newNotFoundError(responseBody)
	}

	if response.StatusCode == 401 {
		return 0, []byte{}, newUnauthorizedError(responseBody)
	}

	for _, acceptableCode := range requestArgs.AcceptableStatusCodes {
		if response.StatusCode == acceptableCode {
			return response.StatusCode, responseBody, nil
		}
	}

	return response.StatusCode, responseBody, newUnexpectedStatusError(response.StatusCode, responseBody)
}

func (client Client) printRequest(request *http.Request) {
	if os.Getenv("TRACE") != "" {
		fmt.Printf("\nREQUEST: %+v\n", request)
	}
}

func (client Client) printResponse(status int, body []byte) {
	if os.Getenv("TRACE") != "" {
		fmt.Printf("\nRESPONSE: %d %s\n", status, body)
	}
}

func (client Client) unmarshal(body []byte, response interface{}) error {
	err := json.Unmarshal(body, response)
	if err != nil {
		return newResponseBodyUnmarshalError(err)
	}
	return nil
}
