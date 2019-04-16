package fstest

import (
	"context"
	"testing"
)

func TestSetupAndTeardown(t *testing.T) {
	projectId1 := "test-app1"
	client, err := Setup(projectId1)
	if err != nil {
		t.Errorf("failed to setup: %v", err)
		return
	}
	if len(defaultContext.instances) != 1 {
		t.Errorf("expected 1, actual: %d", len(defaultContext.instances))
		Teardown(projectId1)
		return
	}

	data := map[string]interface{}{
		"id": 1,
	}
	_, err = client.Collection("users").
		Doc("test").
		Set(context.Background(), data)
	if err != nil {
		t.Errorf("failed to set doc: %v", err)
	}

	Teardown(projectId1)
	if len(defaultContext.instances) != 0 {
		t.Errorf("expected 1, actual: %d", len(defaultContext.instances))
		return
	}
}
