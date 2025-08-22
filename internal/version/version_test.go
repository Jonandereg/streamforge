package version

import "testing"

func TestDefaults(t *testing.T) {
	if Version == "" || Commit == "" || BuildDate == "" {
		t.Fatal("version variables should be set or non-empty defaults")
	}
}
