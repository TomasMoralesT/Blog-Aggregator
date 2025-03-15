package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	feed "github.com/TomasMoralesT/gator/internal/auth"
	"github.com/TomasMoralesT/gator/internal/config"
	"github.com/TomasMoralesT/gator/internal/database"
	_ "github.com/TomasMoralesT/gator/internal/feed"
	"github.com/TomasMoralesT/gator/internal/follow"
	"github.com/TomasMoralesT/gator/internal/middleware"
	"github.com/TomasMoralesT/gator/internal/unfollow"
	_ "github.com/lib/pq"
)

func main() {

	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	dbQueries := database.New(db)

	appState := &config.State{
		Db:            dbQueries,
		Configuration: &cfg,
	}

	cmds := &config.Commands{
		Handlers: make(map[string]func(*config.State, config.Command) error),
	}
	cmds.Register("register", config.HandlerRegister)
	cmds.Register("login", config.HandlerLogin)
	cmds.Register("reset", config.HandlerReset)
	cmds.Register("users", config.HandlerUsers)
	cmds.Register("agg", middleware.MiddlewareLoggedIn(config.HandlerAgg))
	cmds.Register("addfeed", middleware.MiddlewareLoggedIn(feed.AddFeed))
	cmds.Register("feeds", feed.Feeds)
	cmds.Register("follow", middleware.MiddlewareLoggedIn(follow.Follow))
	cmds.Register("following", middleware.MiddlewareLoggedIn(follow.Following))
	cmds.Register("unfollow", middleware.MiddlewareLoggedIn(unfollow.Unfollow))

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Error: Not enough arguments. A command name is required.")
		os.Exit(1)
	}

	cmd := config.Command{
		Name: os.Args[1],
		Args: os.Args[2:],
	}

	if err := cmds.Run(appState, cmd); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
