package unfollow

import (
	"context"
	"errors"
	"fmt"

	"github.com/TomasMoralesT/gator/internal/config"
	"github.com/TomasMoralesT/gator/internal/database"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func Unfollow(state *config.State, cmd config.Command, user database.User) error {
	args := cmd.Args

	if len(args) < 1 {
		return errors.New("usage: unfollow [url]")
	}

	url := args[0]

	feed, err := state.Db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("error fetching feed by URL: %w", err)
	}

	if feed.ID == uuid.Nil {
		return errors.New("no feed found with the given URL")
	}

	params := database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}

	err = state.Db.DeleteFeedFollow(context.Background(), params)
	if err != nil {
		return fmt.Errorf("error unfollowing the feed: %w", err)
	}

	fmt.Println("Successfully unfollowed the feed!")
	return nil
}
