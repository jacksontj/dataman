package integrationtest

import (
	"testing"
)

// First user of the integration tests. The goal here is to create the various nodes required, and pass them to
// the test function to let it do the tests.
func Test_basic(t *testing.T) {
	taskNode, routerNode, datasourceInstance := Setup()

	// Run tests
	RunIntegrationTests(t, taskNode, routerNode, datasourceInstance)
}
