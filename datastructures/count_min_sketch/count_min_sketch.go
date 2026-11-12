package countminsketch

import (
	"fmt"
	"hash/maphash"
	"math"
	"sort"
)

type CountMinSketch struct {
	depth int
	width int
	seeds []maphash.Seed
	store [][]int
}

func New(depth int, width int) *CountMinSketch {
	if depth <= 0 || width <= 0 {
		panic("depth and width must be positive")
	}

	seeds := make([]maphash.Seed, depth)
	store := make([][]int, depth)

	for row := range depth {
		seeds[row] = maphash.MakeSeed()
		store[row] = make([]int, width)
	}

	return &CountMinSketch{
		depth: depth,
		width: width,
		store: store,
		seeds: seeds,
	}
}
func NewWithSeeds(depth int, width int, seeds []maphash.Seed) *CountMinSketch {
	if depth <= 0 || width <= 0 {
		panic("depth and width must be positive")
	}

	if len(seeds) != depth {
		panic("length of seeds must match depth")
	}

	store := make([][]int, depth)

	for row := range depth {
		store[row] = make([]int, width)
	}

	return &CountMinSketch{
		depth: depth,
		width: width,
		store: store,
		seeds: seeds,
	}
}

func (c *CountMinSketch) Insert(item string) {
	for row := range c.depth {
		hash := maphash.String(c.seeds[row], item)
		column := hash % uint64(c.width)
		c.store[row][column]++
	}
}

// TODO:
// Thread Safe Version Of Insert
func (c *CountMinSketch) InsertSafe(item string) {

}

func (c *CountMinSketch) Count(item string) int {
	smallest := math.MaxInt

	for row := range c.depth {
		hash := maphash.String(c.seeds[row], item)
		column := hash % uint64(c.width)

		smallest = min(smallest, c.store[row][column])
	}

	return smallest
}

func (c *CountMinSketch) Clear() {
	for row := range c.depth {
		for column := range c.width {
			c.store[row][column] = 0
		}
	}
}

func (c *CountMinSketch) Merge(other *CountMinSketch) (*CountMinSketch, error) {
	if c.depth != other.depth || c.width != other.width {
		return nil, fmt.Errorf("cannot merge sketches with different dimensions")
	}

	for row := range c.depth {
		if c.seeds[row] != other.seeds[row] {
			return nil, fmt.Errorf("cannot merge sketches with different hash seeds")
		}
	}

	merged := NewWithSeeds(c.depth, c.width, c.seeds)

	for row := range c.depth {
		for column := range c.width {
			merged.store[row][column] = c.store[row][column] + other.store[row][column]
		}
	}

	return merged, nil
}

// Could use Quickselect to avoid sorting all candidates when k is small.
func (c *CountMinSketch) TopK(k int, candidates []string) []string {
	if k > len(candidates) || k <= 0 {
		panic("k is out of bounds")
	}

	type entry struct {
		item  string
		count int
	}

	entries := make([]entry, 0, len(candidates))

	for _, item := range candidates {
		entries = append(entries, entry{
			item:  item,
			count: c.Count(item),
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		// lexicographically sort if counts are equal
		if entries[i].count == entries[j].count {
			return entries[i].item < entries[j].item
		}
		return entries[i].count > entries[j].count
	})

	result := make([]string, k)

	for i := range k {
		result[i] = entries[i].item
	}

	return result
}
