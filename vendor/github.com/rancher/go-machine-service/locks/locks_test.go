package locks

import (
	_ "fmt"
	"testing"
)

func TestLock(t *testing.T) {

	unlocker := Lock("foo1")
	if unlocker == nil {
		t.Errorf("Didn't obtain lock")
	}

	unlocker2 := Lock("foo1")
	if unlocker2 != nil {
		t.Error("Did obtain lock")
	}

	unlocker.Unlock()
	unlocker = Lock("foo1")
	if unlocker == nil {
		t.Errorf("Didn't obtain lock")
	}

}
