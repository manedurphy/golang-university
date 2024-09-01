package main

import (
	"fmt"
	"iter"
	"math/rand"
	"runtime"
	"time"
)

type Course struct {
	ID         int
	Name       string
	University string
}

var (
	courseNames = []string{
		"Chem-1",
		"Chem-2",
		"Physics-1",
		"Physics-2",
		"Physics-3",
		"Calculus-1",
		"Calculus-2",
		"Calculus-3",
	}

	universities = []string{
		"SJSU",
		"SDSU",
		"UCB",
		"UCSF",
	}
)

func generateCourses(numCourses int) iter.Seq[Course] {
	return func(yield func(Course) bool) {
		for i := range numCourses {
			course := Course{
				ID:         i,
				Name:       courseNames[rand.Intn(len(courseNames))],
				University: universities[rand.Intn(len(universities))],
			}

			if !yield(course) {
				return
			}
		}
	}
}

func main() {
	// Memory before generating courses
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	now := time.Now()
	courses := generateCourses(10000000)
	since := time.Since(now)
	fmt.Printf("took %.2f seconds to create iterator\n", since.Seconds())

	now = time.Now()
	for course := range courses {
		course.ID++
	}
	since = time.Since(now)
	fmt.Printf("took %.2f seconds to operate on all courses\n", since.Seconds())

	// Memory after generating courses
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	fmt.Printf("total allocated memory (before): %.2f Mb\n", float64(memBefore.TotalAlloc)/1e6)
	fmt.Printf("total allocated memory (after): %.2f Mb\n", float64(memAfter.TotalAlloc)/1e6)
}
