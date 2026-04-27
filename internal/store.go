package flagapi

import (
	"context"
)

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

func (s *Server) listFlags(ctx context.Context) ([]FlagRecord, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, rule FROM flags ORDER BY name ASC, id ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	flags := make([]FlagRecord, 0)
	for rows.Next() {
		var flag FlagRecord
		if err := rows.Scan(&flag.ID, &flag.Name, &flag.Rule); err != nil {
			return nil, err
		}
		flags = append(flags, flag)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return flags, nil
}


