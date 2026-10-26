package secrand

import (
	"math/rand/v2"
	"testing"

	"github.com/Alievs-corp/hypedomino/server/internal/rules"
)

func TestSourceVaries(t *testing.T) {
	var source Source
	if source.Uint64() == source.Uint64() {
		t.Fatal("two draws returned the same value")
	}
}

func TestSourceDealsValidSets(t *testing.T) {
	random := rand.New(Source{})
	config := rules.DefaultConfig()
	for round := 0; round < 1000; round++ {
		for _, mode := range []rules.Mode{rules.Mode1v1, rules.Mode2v2} {
			deal := rules.NewDeal(mode, config, random)
			if err := rules.ValidateDeal(mode, config, deal); err != nil {
				t.Fatalf("round %d %s: %v", round, mode, err)
			}
		}
	}
}
