package db

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"upsilon_cities_go/lib/misc/config/system"

	// needed for postgres driver
	"github.com/lib/pq"
)

// Handler Contains DB related informations
type Handler struct {
	db   *sql.DB
	open bool
	Name string
	Test bool
}

var testMode bool

//Raw return raw db pointer.
func (dbh *Handler) Raw() *sql.DB {
	return dbh.db
}

//MarkSessionAsTest ensure that all New() call a redirected to NewTest()
func MarkSessionAsTest() {
	testMode = true
}

//New Create a new handler for database, ensure database is created
func New() *Handler {
	if testMode {
		return NewTest()
	}
	handler := new(Handler)
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable host=%s port=%s",
		system.Get("db_user", ""), system.Get("db_password", ""), system.Get("db_name", ""), system.Get("db_host", ""), system.Get("db_port", ""))
	db, _ := sql.Open("postgres", dbinfo)

	errPing := db.Ping()
	if err, ok := errPing.(*pq.Error); ok {
		log.Fatalf("DB: Database failed to be connected: %s", err)
	} else {
		log.Printf("DB: Successfully connected to : %s %s", system.Get("db_host", ""), system.Get("db_name", ""))
	}

	handler.db = db
	handler.open = true
	handler.Name = system.Get("db_name", "")
	handler.Test = false
	return handler
}

//NewTest Create a new handler for test database
func NewTest() *Handler {
	handler := new(Handler)
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable host=%s port=%s",
		system.Get("db_user", ""), system.Get("db_password", ""), system.Get("db_test_name", ""), system.Get("db_host", ""), system.Get("db_port", ""))

	db, _ := sql.Open("postgres", dbinfo)

	errPing := db.Ping()
	if err, ok := errPing.(*pq.Error); ok {
		log.Fatalf("DB: Database failed to be connected: %s", err)
	} else {
		log.Printf("DB: Successfully connected to : %s %s", system.Get("db_host", ""), system.Get("db_test_name", ""))
	}

	handler.db = db
	handler.open = true
	handler.Name = system.Get("db_test_name", "")
	handler.Test = true
	return handler
}

// Exec executes provided query and check if it's correctly executed or not.
// Abort app if not.
// DONT FORGET TO CLOSE RESULT (using result.Close())
func (dbh *Handler) Exec(query string) (result *sql.Rows) {
	dbh.CheckState()
	log.Printf("DB: About to Exec: %s", query)
	result, err := dbh.db.Query(query)
	errorCheck(query, err)
	return result
}

// Query Just like Exec but uses Postgres formater.
// DONT FORGET TO CLOSE RESULT (using result.Close())
func (dbh *Handler) Query(format string, a ...interface{}) (result *sql.Rows) {
	dbh.CheckState()
	log.Printf("DB: About to Query: %s", format)
	result, err := dbh.db.Query(format, a...)
	errorCheck(format, err, a)
	return result
}

// CheckState assert that connection to DB is still alive. or break
func (dbh *Handler) CheckState() {
	if !dbh.open {
		debug.PrintStack()
		log.Fatal("DB: Can't use this connection, it's been closed")
	}
	err := dbh.db.Ping()
	if err != nil {
		debug.PrintStack()
		log.Fatalf("DB: Can't use this connection, an error occured: %s", err)
	}
}

// Close frees db ressource
func (dbh *Handler) Close() {
	if dbh.open {
		dbh.open = false
		defer dbh.db.Close()
	} else {
		log.Print("DB: Already Closed")
	}
}

// ErrorCheck checks if query result has an error or not
func errorCheck(query string, err error, a ...interface{}) bool {
	if err != nil {
		log.Printf("DB: Failed to execute query: %s", query)

		log.Printf("DB: With params: %v", a)
		log.Printf("DB: With error: %s", err)
		// fatal aborts app
		debug.PrintStack()
		log.Fatalf("DB: Aborting: %s", err)

		return true
	}

	return false
}

//CheckVersion Well check migrations and db state and update when necessary.
func CheckVersion(dbh *Handler) {

	dbh.CheckState()
	log.Printf("DB: About to Query: select * from versions")
	result, err := dbh.db.Query("select applied, file from versions order by applied DESC;")

	// ensure last migration date is way in the past.
	applied_migrations := make(map[string]time.Time)
	if err != nil {
		// version table doesn't exist: create database.
		f, ferr := os.Open(system.MakePath(system.Get("db_schema", "db/schema.sql")))
		if ferr != nil {
			log.Fatalln("DB: No schema file found can't initialize database")
		}
		schema, ferr := ioutil.ReadAll(f)
		if ferr != nil {
			log.Fatalln("DB: Schema found but unable to read it all.")
		}
		f.Close()

		q, err := dbh.db.Query(string(schema))

		if err != nil {
			log.Printf("DB: While executing: %s", string(schema))
			log.Fatalf("DB: Unable to apply schema %s ", err)
		}
		q.Close()

		// should insert at least something in version ...

		q, err = dbh.db.Query("insert into versions(file) values ('schema.sql');")
		if err != nil {
			log.Printf("DB: While executing: %s", "insert into versions(file) values ('schema.sql');")
			log.Fatalf("DB: Unable to create versions table %s ", err)
		}

		applied_migrations["schema.sql"] = time.Now().UTC()

		// expect schema to be same as all migrations ... ;)
		err = filepath.Walk(system.MakePath(system.Get("db_migrations", "db/migrations")), func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Fatalf("DB: prevent panic by handling failure accessing a path %q: %v\n", system.Get("db_migrations", "db/migrations"), err)
				return err
			}
			if strings.HasSuffix(info.Name(), ".sql") {

				dbh.Query("insert into versions(file) values ($1);", path).Close()

				applied_migrations[path] = time.Now().UTC()
			}

			return nil
		})
	} else {

		// get lastest applied migration date.
		for result.Next() {
			var applied time.Time
			var file string

			result.Scan(&applied, &file)
			applied_migrations[file] = applied
		}
		result.Close()

	}

	// thus we keep order here ;)
	var orderedFiles []string
	log.Printf("DB: Attempting to find migrations in: %s", system.Get("db_migrations", "db/migrations"))
	err = filepath.Walk(system.MakePath(system.Get("db_migrations", "db/migrations")), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatalf("DB: prevent panic by handling failure accessing a path %q: %v\n", system.MakePath(system.Get("db_migrations", "db/migrations")), err)
			return err
		}
		if strings.HasSuffix(info.Name(), ".sql") {

			migrationFilename := strings.TrimLeft(strings.Replace(path, system.MakePath(system.Get("db_migrations", "db/migrations")), "", 1), string(os.PathSeparator))
			dateString := strings.Split(migrationFilename, "_")[0]
			_, err := time.Parse("200601021504", dateString)

			if err != nil {
				log.Fatalf("DB: One one the migration files: %s, has an invalid format. Expected YYYYMMDDHHMM_name.sql", migrationFilename)
				return err
			}
			log.Printf("DB: Read Migration file: %s", migrationFilename)

			orderedFiles = append(orderedFiles, path)

		}
		return nil
	})

	// ensure file are ordered (they should be by date ;)
	sort.Strings(orderedFiles)

	if err != nil {
		return
	}

	for _, k := range orderedFiles {
		if _, found := applied_migrations[k]; !found {
			log.Printf("DB: Applying migration: %s", k)
			f, ferr := os.Open(k)
			if ferr != nil {
				log.Fatalf("DB: Unable to open migration file %s", k)
			}

			migration, ferr := ioutil.ReadAll(f)
			if ferr != nil {
				log.Fatalf("DB: Unable to read migration file %s", k)
			}

			log.Printf("DB: Applying migration: %s", string(migration))
			q, err := dbh.db.Query(string(migration))

			if err != nil {
				log.Fatalf("DB: Unable to apply migration file %s: %s ", k, err)
			}
			q.Close()

			dbh.Query("insert into versions(file) values ($1);", k).Close()
		}
	}
	log.Printf("DB: DB is up to date ! ")
}

//FlushDatabase Clears everyhting from database and reload it.
func FlushDatabase(dbh *Handler) {
	log.Printf("DB: Flushing Database %s", dbh.Name)
	flush := fmt.Sprintf(`DROP SCHEMA public CASCADE;
						  CREATE SCHEMA public;
						  GRANT ALL ON SCHEMA public TO %s;
						  GRANT ALL ON SCHEMA public TO public;`, system.Get("db_user", ""))

	dbh.Exec(flush).Close()
	CheckVersion(dbh)
}

//ApplySeed seek a seed and apply it to db.
func ApplySeed(dbh *Handler, seed string) error {
	// thus we keep order here ;)
	log.Printf("DB: Attempting to find Seed %s in: %s", seed, system.Get("db_seeds", "db/seeds"))
	return filepath.Walk(system.MakePath(system.Get("db_seeds", "db/seeds")), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatalf("DB: prevent panic by handling failure accessing a path %q: %v\n", system.Get("db_seeds", "db/seeds"), err)
			return err
		}
		if strings.HasSuffix(info.Name(), ".sql") && strings.Contains(info.Name(), seed) {

			log.Printf("DB: Applying seed: %s", info.Name())
			f, ferr := os.Open(path)
			if ferr != nil {
				log.Fatalf("DB: Unable to open seed file %s", path)
			}

			seedContent, ferr := ioutil.ReadAll(f)
			if ferr != nil {
				log.Fatalf("DB: Unable to read seed file %s", path)
			}

			log.Printf("DB: Applying seed: %s", string(seedContent))
			q, err := dbh.db.Query(string(seedContent))

			if err != nil {
				log.Fatalf("DB: Unable to apply seed file %s: %s ", path, err)
			}
			q.Close()
			return nil
		}
		return nil
	})
}
