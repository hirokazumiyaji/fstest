package fstest

import (
	"os"
	"testing"
)

func TestNewInstance(t *testing.T) {
	instance, err := NewInstance(&Options{ProjectId: "test"})
	if err != nil {
		t.Errorf("failed to new instance: %v", err)
	} else {
		if os.Getenv(firestoreHostEnvName) == "" {
			t.Error(firestoreHostEnvName + " is empty")
		}
		err := instance.Close()
		if err != nil {
			t.Errorf("failed to close instance: %v", err)
		}
	}
}
