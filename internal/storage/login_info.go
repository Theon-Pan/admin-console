package storage // import "admin-console/internal/storage"
import (
	"admin-console/internal/model"
	"context"
)

func (s *Storage) InsertLoginInfoContext(ctx context.Context, loginInfo *model.LoginInfoCreationRequest) (info *model.LoginInfo, err error) {
	query := `
		INSERT INTO sys_login_info
			(user_id, username, ip, login_location, browser, os, status, message, login_at)
		VALUES
			($1, LOWER($2), $3, $4, $5, $6, $7, $8, $9)
		RETURNING
			info_id,
			user_id,
			username,
			ip,
			login_location,
			browser,
			os,
			status,
			message,
			login_at
	`
	stmt, err := s.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() // Rolls back on early return; has no effect after Commit.

	txStmt := tx.StmtContext(ctx, stmt)

	info = new(model.LoginInfo)
	err = txStmt.QueryRowContext(
		ctx,
		loginInfo.UserID,
		loginInfo.Username,
		loginInfo.IP,
		loginInfo.LoginLocation,
		loginInfo.Browser,
		loginInfo.OS,
		loginInfo.Status,
		loginInfo.Message,
		loginInfo.LoginAt,
	).Scan(
		&info.InfoID,
		&info.UserID,
		&info.Username,
		&info.IP,
		&info.LoginLocation,
		&info.Browser,
		&info.OS,
		&info.Status,
		&info.Message,
		&info.LoginAt,
	)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return info, nil
}
