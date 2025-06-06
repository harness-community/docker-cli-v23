package container // import "github.com/harness-community/docker-v23/api/types/container"

// ----------------------------------------------------------------------------
// Code generated by `swagger generate operation`. DO NOT EDIT.
//
// See hack/generate-swagger-api.sh
// ----------------------------------------------------------------------------

// ContainerTopOKBody OK response to ContainerTop operation
// swagger:model ContainerTopOKBody
type ContainerTopOKBody struct {

	// Each process running in the container, where each is process
	// is an array of values corresponding to the titles.
	//
	// Required: true
	Processes [][]string `json:"Processes"`

	// The ps column titles
	// Required: true
	Titles []string `json:"Titles"`
}
