package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		currentUser := s.configPtr.Name
		userExists, err := s.db.GetUser(context.Background(), currentUser)
		if err != nil {
			return err
		}
		return handler(s, cmd, userExists)
	}
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

func handleReset(s *state, cmd command) error {
	err := s.db.DeleteUsers(context.Background())
	if err != nil {
		os.Exit(1)
		return err
	}

	fmt.Println("All users deleted successfully")
	return nil
}

func handleGetAllUsers(s *state, cmd command) error {
	currentUser := s.configPtr.Name

	users, err := s.db.GetAllUsers(context.Background())
	if err != nil {
		return err
	}
	for _, user := range users {
		if user == currentUser {
			fmt.Printf("* %v (current)\n", user)
			continue
		} else {
			fmt.Printf("* %v\n", user)
		}
	}
	return nil
}

func handleAgg(s *state, cmd command) error {
	feed, err := config.FetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return err
	}

	fmt.Printf("%v", feed)

	return nil
}

func handleAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		log.Fatal("not enough args")
		os.Exit(1)
	}
	if len(cmd.args) < 4 {
		log.Fatal("not enough args")
		os.Exit(1)
	}
	feedId := uuid.New()
	feedParams := database.CreateFeedParams{
		ID:        feedId,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[2],
		Url:       cmd.args[3],
		UserID:    user.ID,
	}
	newFeed, err := s.db.CreateFeed(context.Background(), feedParams)
	if err != nil {
		os.Exit(1)
		return err
	}

	params := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feedParams.ID,
	}
	follow, err := s.db.CreateFeedFollow(context.Background(), params)
	if err != nil {
		return err
	}

	fmt.Printf("%v Created successfully%v", newFeed.Name, follow.FeedID)

	return nil
}

func handleFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeedsWithUsers(context.Background())
	if err != nil {
		return err
	}

	for _, feed := range feeds {
		fmt.Printf("%v\n%v\n%v\n", feed.Name, feed.Url, feed.Name_2)
	}

	return nil
}

func handleFollow(s *state, cmd command, user database.User) error {
	feedById, err := s.db.GetFeedByUrl(context.Background(), cmd.args[2])
	if err != nil {
		return err
	}
	newId := uuid.New()
	params := database.CreateFeedFollowParams{
		ID:        newId,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feedById,
	}

	follow, err := s.db.CreateFeedFollow(context.Background(), params)
	if err != nil {
		return err
	}

	fmt.Printf("%v\n%v\n", follow.FeedName, follow.UserName)

	return nil
}

func handleFollowing(s *state, cmd command, user database.User) error {
	following, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}

	for _, feed := range following {
		fmt.Printf("%v\n", feed.FeedName)
	}

	return nil
}

func handleUnfollow(s *state, cmd command, user database.User) error {
	params := database.DeleteFeedFollowByURLParams{
		Url:    cmd.args[2],
		UserID: user.ID,
	}

	if err := s.db.DeleteFeedFollowByURL(context.Background(), params); err != nil {
		return err
	}

	return nil
}
