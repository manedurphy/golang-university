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
