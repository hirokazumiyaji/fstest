package fstest

import (
	"context"
	"testing"
)

func TestSetupAndTeardown(t *testing.T) {
	projectId1 := "test-app1"
	client, err := Setup(&Options{ProjectId: projectId1})
	if err != nil {
		t.Errorf("failed to setup: %v", err)
		return
	}
	defer Teardown(projectId1)
	if len(defaultContext.instances) != 1 {
		t.Errorf("expected 1, actual: %d", len(defaultContext.instances))
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

	projectId2 := "test-app2"
	_, err = Setup(&Options{ProjectId: projectId2})
	if err != nil {
		t.Errorf("failed to setup: %v", err)
		return
	}
	defer Teardown(projectId2)
	if len(defaultContext.instances) != 2 {
		t.Errorf("expected 2, actual: %d", len(defaultContext.instances))
		return
	}
	_, err = Setup(&Options{ProjectId: projectId2})
	if err != nil {
		t.Errorf("failed to setup: %v", err)
		return
	}
	defer Teardown(projectId2)
	if len(defaultContext.instances) != 2 {
		t.Errorf("expected 2, actual: %d", len(defaultContext.instances))
		return
	}
	Teardown(projectId1)
	Teardown(projectId2)
	Teardown(projectId2)
	if len(defaultContext.instances) != 0 {
		t.Errorf("expected 0, actual: %d", len(defaultContext.instances))
		return
	}
}
