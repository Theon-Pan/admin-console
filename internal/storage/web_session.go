package storage // import "admin-console/interna/storage"
import (
	"admin-console/internal/model"
	"database/sql"
	"errors"
	"fmt"
)

// CreateWebSession persists a new web session built via model.NewWebSession.
func (s *Storage) CreateWebSession(session *model.WebSession) error {
	if session == nil {
		return errors.New(`store: web session is nil`)
	}

	stateJSON, err := session.MarshalState()
	if err != nil {
		return fmt.Errorf(`store: unable to serializes web session state: %v`, err)
	}

	query := `
		INSERT INTO web_sesions (
			id,
			secret_hash,
			user_agent,
			ip,
			state
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at
	`

	err = s.db.QueryRow(
		query,
		session.ID,
		session.SecretHash,
		session.UserAgent,
		sql.NullString{String: session.IP, Valid: session.IP != ""},
		stateJSON,
	).Scan(&session.CreatedAt)
	if err != nil {
		return fmt.Errorf(`store: unable to create web session: %v`, err)
	}

	return nil
}

// WebSessionByID returns the web session identified by id, or nil if not found.
func (s *Storage) WebSessionByID(sessionID string) (*model.WebSession, error) {
	if sessionID == "" {
		return nil, nil
	}

	row := s.db.QueryRow(`
		SELECT
			id,
			secret_hash,
			user_id,
			created_at,
			user_agent,
			ip,
			state
		FROM
			web_sessions
		WHERE
			id=$1
	`, sessionID)

	session, err := scanWebSession(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch web session: %v`, err)
	}

	return session, nil
}

type webSessionScanner interface {
	Scan(dest ...any) error
}

func scanWebSession(scanner webSessionScanner) (*model.WebSession, error) {
	var session model.WebSession
	var userID sql.NullInt64
	var ip sql.NullString
	var stateRaw []byte

	err := scanner.Scan(
		&session.ID,
		&session.SecretHash,
		&userID,
		&session.CreatedAt,
		&session.UserAgent,
		&ip,
		&stateRaw,
	)
	if err != nil {
		return nil, err
	}

	if ip.Valid {
		session.IP = ip.String
	}

	if err := session.UnmarshalState(stateRaw); err != nil {
		return nil, err
	}

	return &session, nil
}
