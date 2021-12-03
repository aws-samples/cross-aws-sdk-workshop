package workshop

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

// NewTemporaryRedirectResponse returns an API gateway HTTP message of a
// temporary redirect to the target location.
func NewTemporaryRedirectResponse(location string) (*events.APIGatewayV2HTTPResponse, error) {
	return &events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusTemporaryRedirect,
		Headers: map[string]string{
			"location": location,
		},
	}, nil
}

// NewBadRequestErrorResponse returns an API gateway HTTP error response for
// HTTP 400 BadRequest message.
func NewBadRequestErrorResponse(message string) (*events.APIGatewayV2HTTPResponse, error) {
	return NewJSONResponse(400, nil, ErrorMessageResponse{
		Code:    "BadRequestError",
		Message: "BadRequestError: " + message,
	})
}

// NewNotFoundErrorResponse returns an API gateway HTTP error response for
// HTTP 404 NotFound message.
func NewNotFoundErrorResponse(message string) (*events.APIGatewayV2HTTPResponse, error) {
	return NewJSONResponse(404, nil, ErrorMessageResponse{
		Code:    "NotFoundError",
		Message: "NotFoundError: " + message,
	})
}

// NewTooManyRequestsErrorResponse returns an API gateway HTTP error response for
// HTTP 429 TooManyRequests message.
func NewTooManyRequestsErrorResponse(message string) (*events.APIGatewayV2HTTPResponse, error) {
	return NewJSONResponse(429, nil, ErrorMessageResponse{
		Code:    "TooManyRequestsError",
		Message: "TooManyRequestsError: " + message,
	})
}

// NewJSONResponse cosntructs an API Gateway HTTP response value for the parameters
// provided. Serializing the payload as a JSON document in the response.
func NewJSONResponse(status int, header http.Header, payload interface{}) (
	*events.APIGatewayV2HTTPResponse, error,
) {
	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil,
			fmt.Errorf("failed to serialize response, %w", err)
	}
	if header == nil {
		header = http.Header{}
	}
	header.Set("Content-Type", "application/json")

	headers := map[string]string{}
	for k, vs := range header {
		if len(vs) == 0 {
			continue
		}
		headers[strings.ToLower(k)] = vs[0]
	}
	return &events.APIGatewayV2HTTPResponse{
		StatusCode: status,
		Headers:    headers,
		Body:       string(body),
	}, nil
}

// ErrorMessageResponse provides the structured message for API errors sent
// to the client.
type ErrorMessageResponse struct {
	Code    string `json:"Code"`
	Message string `json:"Message"`
}
