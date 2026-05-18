package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/beast447/RSSfeedCLI/internal/config"
	"github.com/beast447/RSSfeedCLI/internal/database"
	"github.com/google/uuid"
)

type state struct {
	db        *database.Queries
	configPtr *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	list map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	command, exists := c.list[cmd.name]
	if !exists {
		return fmt.Errorf("unknown command: %s", cmd.name)
	}
	return command(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.list[name] = f
}

func handleLogin(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("no arguments were given")
	}
	if len(cmd.args) < 3 {
		return fmt.Errorf("no arguments were given after login")
	}

	userExists, err := s.db.GetUser(context.Background(), cmd.args[2])
	if err != nil {
		os.Exit(1)
		return err
	}
	if userExists.Name != cmd.args[2] {
		os.Exit(1)
	}

	if err := config.SetUser(cmd.args[2]); err != nil {
		return err
	}

	s.configPtr.Name = cmd.args[2]
	fmt.Println("User set to", cmd.args[2])
	return nil
}

func handleRegister(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("Not enough args")
	}
	if len(cmd.args) < 3 {
		return fmt.Errorf("Not enough args")
	}

	currentTime := sql.NullTime{Time: time.Now(), Valid: true}
	userID := uuid.New()

	userExists, err := s.db.GetUser(context.Background(), cmd.args[2])
	if err == nil {
		os.Exit(1)
		return err
	}
	if userExists.Name == cmd.args[2] {
		fmt.Println("User exists")
		os.Exit(1)
	}

	user := database.CreateUserParams{ID: userID, CreatedAt: currentTime, UpdatedAt: currentTime, Name: cmd.args[2]}

	_, errr := s.db.CreateUser(context.Background(), user)
	if errr != nil {
		return err
	}

	if err := config.SetUser(cmd.args[2]); err != nil {
		return err
	}

	fmt.Printf("%v created", cmd.args[2])
	return nil
}
