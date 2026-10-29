package main

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"

	"golang.org/x/net/html"
)

func main() {
	var allocations []uint64
	for _, count := range []int{8192, 16384} {
		input := strings.Repeat("a</x>", count)
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		started := time.Now()
		_, err := html.Parse(strings.NewReader(input))
		runtime.ReadMemStats(&after)
		if err != nil { panic(err) }
		encoded, _ := json.Marshal(map[string]any{"runs": count, "input_bytes": len(input), "allocated_bytes": after.TotalAlloc-before.TotalAlloc, "elapsed_nanoseconds": time.Since(started).Nanoseconds()})
		fmt.Println(string(encoded))
		allocations = append(allocations, after.TotalAlloc-before.TotalAlloc)
	}
	ratio := float64(allocations[1]) / float64(allocations[0])
	fmt.Printf("allocation growth: %.3f\n", ratio)
	if ratio > 2.7 { panic("HTML adjacent-text allocation growth is not near-linear") }
}
