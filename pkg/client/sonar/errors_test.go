package sonar

import "testing"

func TestIsErrNotFound(t *testing.T) {
	if IsErrNotFound(nil) {
		t.Fatal("wrong error")
	}
}
