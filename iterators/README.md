---
layout: post
title: Iterators in Golang
tags: golang iterators
---

# Table of Contents
- [Table of Contents](#table-of-contents)
- [Example 1: Database](#example-1-database)
  - [Database](#database)
  - [Main](#main)

# Example 1: Database

Let's explore a practical example where we use iterators to retrieve data from a database.

## Database

The database functionality has been encapsulated within a package named `db`, which is a common practice to enhance code readability. The main point of interest here is the `GetCourses` method, which utilizes the `iter.Seq2` definition to return two values to the consumer: a `Course` object and an `error`.

```go
// Seq2 is an iterator over sequences of pairs of values, most commonly key-value pairs. When called as seq(yield), seq calls yield(k, v) for each pair (k, v) in the sequence, stopping early if yield returns false.
type Seq2[K, V any] func(yield func(K, V) bool)

// GetCourses returns an iterator of Course objects
GetCourses() iter.Seq2[Course, error]
```

This method begins by querying all rows. If an error occurs, it is returned to the consumer, and the iteration stops since there is nothing to iterate through. As we process each row, note that we do not return immediately upon encountering an error; instead, we yield the error back to the caller and proceed to the next row. The decision to continue or stop depends on your application's requirements. This example demonstrates that encountering an error while scanning a row does not necessarily force a return.

```go
package db

import (
	"database/sql"
	"fmt"
	"iter"
	"math/rand"

	_ "github.com/mattn/go-sqlite3"
)

type (
	CoursesDB interface {
		// Seed seeds the database with the number of courses specified
		Seed(numCourses int) error

		// GetCourses returns an iterator of Course objects
		GetCourses() iter.Seq2[Course, error]

		// Close closes the database
		Close() error
	}

	Course struct {
		ID         int
		Name       string
		University string
	}

	coursesDB struct {
		db *sql.DB
	}
)

const (
	selectSQL    = `SELECT * FROM courses`
	insertSQL    = `INSERT INTO courses(name, university) VALUES (?, ?)`
	dropTableSQL = `DROP TABLE IF EXISTS courses`

	createTableSQL = `CREATE TABLE IF NOT EXISTS courses (
        "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,   
        "name" TEXT,
        "university" TEXT
    );`
)

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

// New creates a new CoursesDB instance
func New(dataDir string) (CoursesDB, error) {
	var (
		db  *sql.DB
		err error
	)

	// Open database
	db, err = sql.Open("sqlite3", fmt.Sprintf("%s/courses.db", dataDir))
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &coursesDB{
		db: db,
	}, nil
}

func (d *coursesDB) Seed(numCourses int) error {
	var (
		tx        *sql.Tx
		statement *sql.Stmt
		err       error
	)

	_, err = d.db.Exec(dropTableSQL)
	if err != nil {
		return fmt.Errorf("failed to drop table: %w", err)
	}

	_, err = d.db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	tx, err = d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	statement, err = tx.Prepare(insertSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare SQL statment: %w", err)
	}
	defer statement.Close()

	// Seed database
	for course := range d.generateCourses(numCourses) {
		_, err = statement.Exec(course.Name, course.University)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to prepare SQL statment: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (d *coursesDB) GetCourses() iter.Seq2[Course, error] {
	return func(yield func(Course, error) bool) {
		var (
			rows *sql.Rows
			err  error
		)

		rows, err = d.db.Query(selectSQL)
		if err != nil {
			// When an error is encountered, we should yield it back to
			// the consumer an stop the iterator
			yield(Course{}, err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var c Course

			err = rows.Scan(&c.ID, &c.Name, &c.University)
			if !yield(c, err) {
				return
			}
		}

		err = rows.Err()
		if err != nil {
			// When an error is encountered, we should yield it back to
			// the consumer an stop the iterator
			yield(Course{}, err)
			return
		}
	}
}

func (d *coursesDB) Close() error {
	return d.db.Close()
}

// Generator of Course objects
func (d *coursesDB) generateCourses(numCourses int) iter.Seq[Course] {
	return func(yield func(Course) bool) {
		for range numCourses {
			course := Course{
				Name:       courseNames[rand.Intn(len(courseNames))],
				University: universities[rand.Intn(len(universities))],
			}

			if !yield(course) {
				return
			}
		}
	}
}
```

## Main

With our database abstraction layer, the main program becomes very easy to read. It simply creates a new `CoursesDB` instance, seeds it with a specified number of courses, and then queries and iterates through all the data via the iterator returned by `GetCourses`.

```go
package main

import (
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/manedurphy/golang-university/iterators/01-database/db"
)

var (
	dataDir    string
	numCourses int
)

func init() {
	flag.StringVar(&dataDir, "data-dir", "", "The directory for storing the DB file")
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

	// Get courses from database using iterator
	for course, err := range coursesDB.GetCourses() {
		if err != nil {
			logger.Error("failed to get course", "err", err)
			continue
		}

		logger.Info("received course", "course", course)
	}
}
```
