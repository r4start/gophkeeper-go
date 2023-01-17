package migrations

const (
	_createUserTable = `create table if not exists user (
    	user_id text primary key,
    	token text,
    	refresh_token text,
    	master_key blob,
    	key_salt blob,
		last_sync_ts bigint
    );`

	_createFilesTable = `create table if not exists files (
		id text primary key,
		user_id text,
		name text,
		path text,
		key blob,
		salt blob,
		added_ts bigint
	);`

	_createCardsTable = `create table if not exists cards (
		id text primary key,
		user_id text,
		name text,
		number text,
		holder text,
		security_code text,
		expiry_date text,
		added_ts bigint
	);`

	_createCredentialsTable = `create table if not exists creds (
		id text primary key,
		user_id text,
		username text,
		password text,
		uri text,
		description text,
		added_ts bigint
	);`
)

var Migrations []string

func init() {
	Migrations = append(Migrations,
		_createUserTable, _createFilesTable,
		_createCardsTable, _createCredentialsTable)
}
