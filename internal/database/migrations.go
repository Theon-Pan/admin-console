package database // import "admin-console/internal/database"
import "database/sql"

var schemaVersion = len(migrations)

// Order is important. Add new migrations at the end of the list.
var migrations = [...]func(tx *sql.Tx) error{
	func(tx *sql.Tx) (err error) {
		sql := `
			CREATE TABLE sys_schema_version (
				version text not null
			);

			CREATE TABLE sys_users (
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
			
			CREATE INDEX sys_users_user_id_idx
				ON sys_users (user_id);
			
			CREATE INDEX sys_users_username_idex
				ON sys_users (username);
		`
		_, err = tx.Exec(sql)
		return err
	},
	func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(`
			CREATE TABLE sys_web_sessions (
				id text not null,
				secret_hash bytea not null,
				user_id bigint references sys_users(user_id) on delete cascade,
				created_at timestamp with time zone not null default now(),
				user_agent text not null default '',
				ip inet,
				state jsonb not null default '{}'::jsonb,
				primary key (id),
				check (jsonb_typeof(state) = 'object')
			);

			CREATE INDEX sys_web_sessions_user_id_idx
				ON sys_web_sessions (user_id)
				WHERE user_id IS NOT NULL;

			CREATE INDEX sys_web_sessions_created_at_idx
				ON sys_web_sessions (created_at);
		`)
		return err
	},
	func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(`
			CREATE TABLE sys_login_info (
				info_id SERIAL,
				user_id bigint not null,
				username text not null,
				ip inet,
				login_location text,
				browser text,
				os text,
				status varchar(1) not null,
				message text,
				login_at timestamp with time zone not null,
				primary key (info_id)
			);
			
			CREATE INDEX sys_login_info_user_id_idx
				ON sys_login_info (user_id);
			
			CREATE INDEX sys_login_info_username_idx
				ON sys_login_info (username);
			
			CREATE INDEX sys_login_info_login_at_idx
				ON sys_login_info (login_at);
		`)
		return err
	},
}
