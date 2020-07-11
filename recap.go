package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"os"
	"path/filepath"
	"recap/internal"
)

var recapDir, recapDB string

func init() {
	if val, set := os.LookupEnv("RECAP_DIR"); set {
		recapDir = val
	} else {
		if dir, err := os.UserHomeDir(); err != nil {
			panic(err)
		} else {
			recapDir = filepath.Join(dir, "recap")
			fmt.Fprintf(os.Stderr, "RECAP_DIR not set. Defaulting to \"%s\"\n", recapDir)
		}
	}

	if val, set := os.LookupEnv("RECAP_DB"); set {
		recapDB = val
	} else {
		recapDB = "recap"
		fmt.Fprintf(os.Stderr, "RECAP_DB not set. Defaulting to \"%s\"\n", recapDB)
	}
}

func main() {
	connection := fmt.Sprintf("dbname=%s host=/tmp sslmode=disable", recapDB)
	db, err := sql.Open("postgres", connection)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	cli := new(internal.CLI)
	cli.Initialize(db, recapDir)

	cli.Start()
}
