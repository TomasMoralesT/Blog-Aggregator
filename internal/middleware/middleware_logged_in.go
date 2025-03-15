package middleware

import (
	"context"
	"errors"
	"fmt"

	"github.com/TomasMoralesT/gator/internal/config"
	"github.com/TomasMoralesT/gator/internal/database"
)

func MiddlewareLoggedIn(handler func(s *config.State, cmd config.Command, user database.User) error) func(*config.State, config.Command) error {

	return func(s *config.State, cmd config.Command) error {
		if s.Configuration.CurrentUserName == "" {
			return errors.New("no user is logged in")
		}

		user, err := s.Db.GetUserByName(context.Background(), s.Configuration.CurrentUserName)
		if err != nil {
			return fmt.Errorf("error fetching user: %w", err)
		}
		return handler(s, cmd, user)
	}
}
