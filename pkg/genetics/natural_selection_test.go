package genetics_test

import (
	"math/rand"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/inlined/xkcdrand"

	"github.com/inlined/goldmine/pkg/genetics"
)

func TestStochasticUniversalSampling(t *testing.T) {
	for _, test := range []struct {
		tag             string
		numParents      int
		source          rand.Source
		fitness         []genetics.Fitness
		expectedIndexes []int
	}{
		{
			tag:             "pick every other (even)",
			numParents:      3,
			source:          xkcdrand.Sequence(1),
			fitness:         []genetics.Fitness{2, 2, 2, 2, 2, 2},
			expectedIndexes: []int{0, 2, 4},
		},
		{
			tag:             "pick every other (odd)",
			numParents:      3,
			source:          xkcdrand.Sequence(3),
			fitness:         []genetics.Fitness{2, 2, 2, 2, 2, 2},
			expectedIndexes: []int{1, 3, 5},
		},
		{
			// This is an edge case and a major sign to switch the selection mechanism to ranked scoring
			// TODO: should this also return a diversity score to help automate the scoring algorithm?
			tag:             "top-exclusively",
			numParents:      3,
			source:          xkcdrand.Sequence(1),
			fitness:         []genetics.Fitness{10, 1, 1},
			expectedIndexes: []int{0, 0, 0},
		},
		{
			tag:             "redundant picks",
			numParents:      3,
			source:          xkcdrand.Sequence(2),
			fitness:         []genetics.Fitness{10, 1, 1},
			expectedIndexes: []int{0, 0, 1},
		},
	} {
		t.Run(test.tag, func(t *testing.T) {
			s := genetics.StochasticUniversalSampling(test.source)
			got := s.SelectParents(test.numParents, test.fitness)
			if diff := cmp.Diff(got, test.expectedIndexes); diff != "" {
				t.Fatalf("Got wrong indexes; got=%v; want=%v; diff=%v", got, test.expectedIndexes, diff)
			}
		})
	}
}
