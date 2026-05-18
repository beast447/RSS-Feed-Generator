package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/beast447/RSSfeedCLI/internal/config"
	"github.com/beast447/RSSfeedCLI/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	data, err := config.Read()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error reading config:", err)
		os.Exit(1)
	}

	s := state{configPtr: &data}
	db, err := sql.Open("postgres", s.configPtr.Db_url)
	if err != nil {
		log.Fatal("Error opening a connection to database")
	}

	s = state{configPtr: &data, db: database.New(db)}
	c := commands{
		list: make(map[string]func(*state, command) error),
	}

	c.register("login", handleLogin)
	c.register("register", handleRegister)
	c.register("reset", handleReset)
	c.register("users", handleGetAllUsers)

	userArgs := os.Args
	if len(userArgs) < 1 {
		fmt.Fprintln(os.Stderr, "not enough arguments provided")
		os.Exit(1)
	}

	cmd := command{name: userArgs[1], args: userArgs}

	if err := c.run(&s, cmd); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
