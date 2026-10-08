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
}
