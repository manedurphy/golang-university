package main

import (
	"flag"
	"iter"
	"log/slog"
	"os"
	"time"

	"github.com/manedurphy/golang-university/iterators/04-database/db"
)

var (
	dataDir    string
	numCourses int
)

func init() {
	flag.StringVar(&dataDir, "data-dir", ".", "The directory for storing the DB file")
	flag.IntVar(&numCourses, "num-courses", 0, "The number of courses to create")
}

func main() {
	var (
		coursesDB db.CoursesDB
		now       time.Time
		logger    *slog.Logger
		err       error
	)

	flag.Parse()

	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create new database instance
	coursesDB, err = db.New(dataDir)
	if err != nil {
		logger.Error("failed to create database", "err", err)
		os.Exit(1)
	}
	defer coursesDB.Close()

	now, err = time.Now(), coursesDB.Seed(numCourses)
	if err != nil {
		logger.Error("failed to seed database", "err", err)
		os.Exit(1)
	}
	logger.Info("successfully seeded database", "duration_ms", time.Since(now).Milliseconds())

	next, stop := iter.Pull2(coursesDB.GetCourses())
	defer stop()

	// Get courses from database using iterator
	for {
		course, err, valid := next()
		if !valid {
			logger.Info("iteration has completed")
			break
		}

		if err != nil {
			logger.Error("failed to get course", "err", err)
			continue
		}

		logger.Info("received course", "course", course)
	}
}
