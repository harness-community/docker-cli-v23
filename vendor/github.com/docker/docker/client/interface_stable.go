package client // import "github.com/DevanshMathur19/docker-v23/client"

// APIClient is an interface that clients that talk with a docker server must implement.
type APIClient interface {
	CommonAPIClient
	apiClientExperimental
}

// Ensure that Client always implements APIClient.
var _ APIClient = &Client{}
