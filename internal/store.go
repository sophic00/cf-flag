package flagapi

import (
	"context"
	"strings"
)

func (s *Server) insertUser(ctx context.Context, user UserRecord) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO users (id, name, email, country) VALUES (?, ?, ?, ?)`,
		user.ID, user.Name, user.Email, user.Country,
	)
	return err
}

func (s *Server) insertFlag(ctx context.Context, flag FlagRecord) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO flags (id, name, rule) VALUES (?, ?, ?)`,
		flag.ID, flag.Name, flag.Rule,
	)
	return err
}

func (s *Server) getFlagRule(ctx context.Context, flagID string) (string, error) {
	var rule string
	err := s.db.QueryRowContext(ctx, `SELECT rule FROM flags WHERE id = ?`, flagID).Scan(&rule)
	if err != nil {
		return "", err
	}
	return rule, nil
}

func (s *Server) getUserCountry(ctx context.Context, userID string) (string, error) {
	var country string
	err := s.db.QueryRowContext(ctx, `SELECT country FROM users WHERE id = ?`, userID).Scan(&country)
	if err != nil {
		return "", err
	}
	return strings.ToUpper(strings.TrimSpace(country)), nil
}
