package rainmaker

func NewRequestArguments(method, path, token string, body interface{}) requestArguments {
	return requestArguments{
		Method: method,
		Path:   path,
		Token:  token,
		Body:   body,
	}
}

func (client Client) MakeRequest(requestArgs requestArguments) (int, []byte, error) {
	return client.makeRequest(requestArgs)
}

func (client Client) Unmarshal(body []byte, response interface{}) error {
	return client.unmarshal(body, response)
}
