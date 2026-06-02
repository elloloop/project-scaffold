package serverkit

import (
	"testing"

	"github.com/elloloop/project-scaffold/packages/go/testkit"
)

func TestHealth(t *testing.T) {
	response := Health("project_scaffold")

	testkit.Equal(t, response.Service, "Project Scaffold")
	testkit.Equal(t, response.Status, "ok")
}
