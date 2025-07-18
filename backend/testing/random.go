package utils_test

import (
	"math/rand"
)

type RandomValueGenerator struct {
	r *rand.Rand
}

func NewRanomValueGenerator(seedValue int64) *RandomValueGenerator {

	r := rand.New(rand.NewSource(seedValue))
	return &RandomValueGenerator{r: r}
}

func (g *RandomValueGenerator) GenerateRandomValue(min, max int) int {
	return g.r.Intn(max-min) + min
}
