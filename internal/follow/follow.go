package follow

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/TomasMoralesT/gator/internal/config"
	"github.com/TomasMoralesT/gator/internal/database"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func Follow(state *config.State, cmd config.Command, user database.User) error {
	args := cmd.Args

	if len(args) < 1 {
		return errors.New("usage: follow [url]")
	}

	url := args[0]

	feed, err := state.Db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("error fetching feed by URL: %w", err)
	}

	if feed.ID == uuid.Nil {
		return errors.New("no feed found with the given URL")
	}

	now := time.Now().UTC()
	feedFollowParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		UserID:    user.ID,
		FeedID:    feed.ID,
	}

	followRows, err := state.Db.CreateFeedFollow(context.Background(), feedFollowParams)
	if err != nil {
		return fmt.Errorf("error creating feed follow: %w", err)
	}

	if len(followRows) == 0 {
		return errors.New("no follow record was created")
	}

	follow := followRows[0]

	fmt.Printf("Feed '%s' successfully followed by user '%s'.\n", follow.FeedName, follow.UserName)
	return nil

}
