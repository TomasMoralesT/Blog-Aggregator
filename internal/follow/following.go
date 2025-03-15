package follow

import (
	"context"

	"fmt"

	"github.com/TomasMoralesT/gator/internal/config"
	"github.com/TomasMoralesT/gator/internal/database"
	_ "github.com/lib/pq"
)

func Following(state *config.State, cmd config.Command, user database.User) error {

	feedFollows, err := state.Db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("error fetching feed follows for user: %w", err)
	}

	if len(feedFollows) == 0 {
		fmt.Printf("User '%s' is not following any feeds.\n", user.Name)
		return nil
	}

	fmt.Printf("User '%s' is following these feeds:\n", user.Name)
	for _, follow := range feedFollows {
		fmt.Printf("- Feed: '%s' (URL: %s)\n", follow.FeedName, follow.FeedUrl)
	}

	return nil
}
