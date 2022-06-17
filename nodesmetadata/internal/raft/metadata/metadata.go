package metadata

const HTTPPath = "/metadata"

type Response struct {
	ApplicationAddress string `json:"application_address"`
}
