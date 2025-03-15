package config

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/TomasMoralesT/gator/internal/database"
	rss "github.com/TomasMoralesT/gator/internal/feed"
	"github.com/google/uuid"
)

type Config struct {
	DBURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

type State struct {
	Db            *database.Queries
	Configuration *Config
}

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	Handlers map[string]func(*State, Command) error
}

const configFileName = ".gatorconfig.json"

func Read() (Config, error) {
	var cfg Config

	configPath, err := getConfigFilePath()
	if err != nil {
		return cfg, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return cfg, err
	}

	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, configFileName), nil
}

func write(cfg Config) error {
	jsonData, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	configPath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, jsonData, 0644)
}

func (c *Config) SetUser(username string) error {
	c.CurrentUserName = username
	return write(*c)
}

func SaveConfig(cfg Config) error {
	configFilePath, err := getConfigFilePath()

	if err != nil {
		return fmt.Errorf("failed to get config file path: %w", err)
	}

	file, err := os.Create(configFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(cfg); err != nil {
		return err
	}
	return nil
}

func HandlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("username is required")
	}

	username := cmd.Args[0]

	_, err := s.Db.GetUser(context.Background(), username)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: User %s does not exist\n", username)
		os.Exit(1)
		return err
	}

	s.Configuration.CurrentUserName = username

	if err := SaveConfig(*s.Configuration); err != nil {
		return fmt.Errorf("Failed to save config: %v", err)
	}

	fmt.Printf("User set to %s\n", username)

	return nil
}

func HandlerRegister(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("username is required")
	}

	username := cmd.Args[0]

	user, err := s.Db.CreateUser(
		context.Background(),
		database.CreateUserParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      username,
		},
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create user: %v\n", err)
		os.Exit(1)
		return err
	}

	s.Configuration.CurrentUserName = username

	if err := SaveConfig(*s.Configuration); err != nil {
		return fmt.Errorf("Failed to save config: %v", err)
	}

	fmt.Printf("User %s created successfully\n", username)
	fmt.Printf("User details: %+v\n", user)

	return nil
}

func newState(cfg *Config) *State {
	return &State{Configuration: cfg}
}

func (c *Commands) Register(name string, f func(*State, Command) error) {
	if c.Handlers == nil {
		c.Handlers = make(map[string]func(*State, Command) error)
	}
	c.Handlers[name] = f
}

func (c *Commands) Run(s *State, cmd Command) error {
	handler, exists := c.Handlers[cmd.Name]
	if !exists {
		return fmt.Errorf("unknown command: %s", cmd.Name)
	}
	return handler(s, cmd)
}

func HandlerReset(s *State, cmd Command) error {

	ctx := context.Background()

	err := s.Db.Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to reset database: %v", err)
	}

	fmt.Println("Database reset successfully.")
	return nil
}

func HandlerUsers(s *State, cmd Command) error {
	ctx := context.Background()

	allusers, err := s.Db.GetUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get users %v", err)
	}

	var userlist []string

	for _, username := range allusers {
		if username.Name == s.Configuration.CurrentUserName {
			userlist = append(userlist, (username.Name + " (current)"))
		} else {
			userlist = append(userlist, username.Name)
		}
	}

	for _, user := range userlist {
		fmt.Println("* " + user)

	}
	return nil
}

func HandlerAgg(appState *State, cmd Command, user database.User) error {

	if len(cmd.Args) != 1 {
		return errors.New("expected 1 argument: time_between_reqs")
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.Args[0])

	if err != nil {
		return fmt.Errorf("invalid duration format :%w", err)
	}

	fmt.Printf("Collecting feeds every %v\n", timeBetweenRequests)

	ticker := time.NewTicker(timeBetweenRequests)
	defer ticker.Stop()

	for ; ; <-ticker.C {
		err := Scrapefeeds(appState)
		if err != nil {
			fmt.Printf("Error scraping feeds: %v\n", err)
		}
	}
	return nil
}

func Scrapefeeds(state *State) error {

	if state == nil {
		return errors.New("invalid state: nil")
	}

	queries := state.Db
	if queries == nil {
		return errors.New("database queries not initialized")
	}

	ctx := context.Background()

	feed, err := queries.GetNextFeedToFetch(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no feeds available to fetch")
		}
		return fmt.Errorf("error getting next feed: %w", err)
	}

	err = queries.MarkFeedFetched(ctx, feed.ID)
	if err != nil {
		return fmt.Errorf("error marking feed as fetched: %w", err)
	}

	fmt.Printf("Fetching feed: %s (%s)\n", feed.Name, feed.Url)

	rssFeed, err := rss.FetchFeed(ctx, feed.Url) // Assuming you have a FetchFeed function
	if err != nil {
		return fmt.Errorf("error fetching feed %s: %w", feed.Url, err)
	}

	fmt.Printf("Found %d items in feed: %s\n", len(rssFeed.Channel.Item), feed.Name)
	for _, item := range rssFeed.Channel.Item {
		fmt.Printf("- %s\n", item.Title)
	}

	return nil
}
