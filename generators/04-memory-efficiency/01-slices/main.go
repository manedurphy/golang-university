package main

import (
	"fmt"
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

func generateCourses(numCourses int) []Course {
	var courses []Course

	for i := range numCourses {
		courses = append(courses, Course{
			ID:         i,
			Name:       courseNames[rand.Intn(len(courseNames))],
			University: universities[rand.Intn(len(universities))],
		})
	}

	return courses
}

func main() {
	// Memory before generating courses
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	now := time.Now()
	courses := generateCourses(10000000)
	fmt.Printf("took %.2f seconds to generate slice\n", time.Since(now).Seconds())

	now = time.Now()
	for _, course := range courses {
		course.ID++
	}
	fmt.Printf("took %.2f seconds to operate on all courses\n", time.Since(now).Seconds())

	// Memory after generating courses
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	fmt.Printf("total allocated memory (before): %.2f Mb\n", float64(memBefore.TotalAlloc)/1e6)
	fmt.Printf("total allocated memory (after): %.2f Mb\n", float64(memAfter.TotalAlloc)/1e6)
}
