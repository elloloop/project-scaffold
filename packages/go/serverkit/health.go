package serverkit

import "github.com/elloloop/project-scaffold/packages/go/platform"

type HealthResponse struct {
	Service string `json:"service"`
	Status  string `json:"status"`
}

func Health(service string) HealthResponse {
	return HealthResponse{
		Service: platform.DisplayName(service),
		Status:  "ok",
	}
}
