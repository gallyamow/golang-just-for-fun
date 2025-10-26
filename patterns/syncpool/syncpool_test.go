package syncpool

import (
	"testing"
)

func TestUsingSyncPool(t *testing.T) {
	instances := UsingSyncPool(10)

	for i := 0; i < len(instances)-1; i++ {
		if instances[i] != instances[i+1] {
			t.Fatalf("used several instances")
		}
	}
}
