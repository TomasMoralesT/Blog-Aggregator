package feed

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/TomasMoralesT/gator/internal/config"
	"github.com/TomasMoralesT/gator/internal/database"
	_ "github.com/lib/pq"
)

func AddFeed(state *config.State, cmd config.Command, user database.User) error {
	args := cmd.Args

	if len(args) < 2 {
		return errors.New("usage: addfeed [name] [url]")
	}

	feedName, feedUrl := args[0], args[1]

	existingFeed, err := state.Db.GetFeedByURL(context.Background(), feedUrl)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("error checking feed existence: %w", err)
	}

	var feedID uuid.UUID
	var dbFeedName string

	if errors.Is(err, sql.ErrNoRows) {
		newFeedParams := database.CreateFeedParams{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			Name:      feedName,
			Url:       feedUrl,
			UserID:    user.ID,
		}
		createdFeed, err := state.Db.CreateFeed(context.Background(), newFeedParams)
		if err != nil {
			return fmt.Errorf("error creating feed: %w", err)
		}
		feedID = createdFeed.ID
		dbFeedName = createdFeed.Name

	} else {
		feedID = existingFeed.ID
		dbFeedName = existingFeed.Name
	}

	feedFollowParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    feedID,
	}

	feedFollowRows, err := state.Db.CreateFeedFollow(context.Background(), feedFollowParams)
	if err != nil {
		return fmt.Errorf("error creating feed follow: %w", err)
	}

	if len(feedFollowRows) == 0 {
		return errors.New("unable to create feed follow record")
	}

	feedFollow := feedFollowRows[0]

	fmt.Printf("Added feed '%s' with URL '%s'.\n", dbFeedName, feedUrl)
	fmt.Printf("User '%s' is now following feed '%s'.\n", feedFollow.UserName, feedFollow.FeedName)

	return nil
}

func Feeds(state *config.State, cmd config.Command) error {
	feeds, err := state.Db.GetAllFeedsWithCreators(context.Background())
	if err != nil {
		return err
	}

	for _, feed := range feeds {
		fmt.Printf("Feed: %s\nURL: %s\n Created by: %s\n\n", feed.FeedName, feed.Url, feed.UserName)
	}
	return nil
}
