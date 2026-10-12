package database // import "admin-console/internal/database"
import "database/sql"

var schemaVersion = len(migrations)

// Order is important. Add new migrations at the end of the list.
var migrations = [...]func(tx *sql.Tx) error{
	func(tx *sql.Tx) (err error) {
		sql := `
			CREATE TABLE schema_version (
				version text not null
			);

			CREATE TABLE users (
				id SERIAL,
				user_id bigint not null unique,
				username text not null unique,
				nickname text,
				email text,
				phonenumber text,
				password text,
				is_admin bool default 'f',
				last_login_at timestamp with time zone,
				remark text,
				primary key (id)
			);
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(`
			CREATE TABLE web_sessions (
				id text not null,
				secret_hash bytea not null,
				user_id bigint references users(user_id) on delete cascade,
				created_at timestamp with time zone not null default now(),
				user_agent text not null default '',
				ip inet,
				state jsonb not null default '{}'::jsonb,
				primary key (id),
				check (jsonb_typeof(state) = 'object')
			);

			CREATE INDEX web_sessions_user_id_idx
				ON web_sessions (user_id)
				WHERE user_id IS NOT NULL;

			CREATE INDEX web_sessions_created_at_idx
				ON web_sessions (created_at);
		`)
		return err
	},
}
