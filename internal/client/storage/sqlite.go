package storage

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/r4start/goph-keeper/internal/client/migrations"
)

const (
	_insertUserData = `insert into user(user_id, token, refresh_token, master_key, key_salt, last_sync_ts)
						values (?, ?, ?, ?, ?, ?)
						on conflict(user_id) do update set
							token = excluded.token,
							refresh_token = excluded.refresh_token,
							master_key = excluded.master_key,
							key_salt = excluded.key_salt,
							last_sync_ts = excluded.last_sync_ts;`

	_getUserData = `select user_id, token, refresh_token, master_key, key_salt from user;`

	_insertFile = `insert into files(id, user_id, name, path, key, salt, added_ts) values (?, ?, ?, ?, ?, ?, ?);`
	_listFiles  = `select id, user_id, name, path, key, salt from files;`
	_selectFile = `select id, user_id, name, path, key, salt from files where id = ?;`
	_deleteFile = `delete from files where id = ?;`

	_insertCard = `insert into cards(id, user_id, name, number, holder, security_code, expiry_date, added_ts) values (?, ?, ?, ?, ?, ?, ?, ?);`
	_listCards  = `select id, number, holder, security_code, expiry_date from cards;`
	_selectCard = `select id, user_id, number, holder, security_code, expiry_date from cards where id = ?;`
	_deleteCard = `delete from cards where id = ?;`

	_insertCred = `insert into creds(id, user_id, username, password, uri, description, added_ts) values (?, ?, ?, ?, ?, ?, ?);`
	_listCreds  = `select id, user_id, username, password, uri, description from creds;`
	_selectCred = `select id, user_id, username, password, uri, description from creds where id = ?;`
	_deleteCred = `delete from creds where id = ?;`
)

var (
	_ Storage = (*sqliteStorage)(nil)
)

type sqliteStorage struct {
	db *sql.DB
}

func NewLocalStorage(dsn string) (*sqliteStorage, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	s := &sqliteStorage{db: db}
	err = s.prepareDB(context.Background(), migrations.Migrations...)
	return s, err
}

func (s *sqliteStorage) UserData(ctx context.Context) (*UserData, error) {
	userData := &UserData{}
	err := s.db.QueryRowContext(ctx, _getUserData).
		Scan(&userData.UserID, &userData.Token, &userData.RefreshToken,
			&userData.MasterKey, &userData.Salt)
	if err != nil {
		return nil, err
	}
	return userData, nil
}

func (s *sqliteStorage) SetUserData(ctx context.Context, ud *UserData) error {
	_, err := s.db.ExecContext(ctx, _insertUserData, ud.UserID, ud.Token,
		ud.RefreshToken, ud.MasterKey, ud.Salt, time.Now().UTC().Unix())
	return err
}

func (s *sqliteStorage) AddFile(ctx context.Context, fd *FileData) error {
	_, err := s.db.ExecContext(ctx, _insertFile, fd.ID, fd.UserID,
		fd.Name, fd.Path, fd.Key.Key, fd.Key.Salt, time.Now().UTC().Unix())
	return err
}

func (s *sqliteStorage) ListFiles(ctx context.Context) ([]FileData, error) {
	rows, err := s.db.QueryContext(ctx, _listFiles)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	files := make([]FileData, 0, 128)
	for rows.Next() {
		var f FileData
		if err := rows.Scan(&f.ID, &f.UserID, &f.Name, &f.Path, &f.Key.Key, &f.Key.Salt); err != nil {
			return nil, err
		}

		files = append(files, f)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}

func (s *sqliteStorage) DeleteFile(ctx context.Context, fd *FileData) error {
	_, err := s.db.ExecContext(ctx, _deleteFile, fd.ID)
	return err
}

func (s *sqliteStorage) FileData(ctx context.Context, id string) (*FileData, error) {
	row := s.db.QueryRowContext(ctx, _selectFile, id)

	data := &FileData{}
	err := row.Scan(&data.ID, &data.UserID, &data.Name, &data.Path, &data.Key.Key, &data.Key.Salt)
	if err != nil {
		return nil, err
	}

	if err := row.Err(); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *sqliteStorage) AddCard(ctx context.Context, data *CardData) error {
	_, err := s.db.ExecContext(ctx, _insertCard, data.ID, data.UserID, data.Name, data.Number,
		data.Holder, data.SecurityCode, data.ExpiryDate, time.Now().UTC().Unix())
	return err
}

func (s *sqliteStorage) ListCards(ctx context.Context) ([]CardData, error) {
	rows, err := s.db.QueryContext(ctx, _listCards)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	cards := make([]CardData, 0)
	for rows.Next() {
		var card CardData
		if err := rows.Scan(&card.ID, &card.Number, &card.Holder,
			&card.SecurityCode, &card.ExpiryDate); err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}

	return cards, nil
}

func (s *sqliteStorage) DeleteCard(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, _deleteCard, id)
	return err
}

func (s *sqliteStorage) CardData(ctx context.Context, id string) (*CardData, error) {
	card := &CardData{}
	err := s.db.QueryRowContext(ctx, _selectCard, id).
		Scan(&card.ID, &card.UserID, &card.Number, &card.Holder,
			&card.SecurityCode, &card.ExpiryDate)
	return card, err
}

func (s *sqliteStorage) AddCredentials(ctx context.Context, data *CredentialData) error {
	_, err := s.db.ExecContext(ctx, _insertCred, data.ID, data.UserID,
		data.Username, data.Password, data.Uri, data.Description, time.Now().UTC().Unix())
	return err
}

func (s *sqliteStorage) ListCredentials(ctx context.Context) ([]CredentialData, error) {
	rows, err := s.db.QueryContext(ctx, _listCreds)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	result := make([]CredentialData, 0)
	for rows.Next() {
		var creds CredentialData
		if err := rows.Scan(&creds.ID, &creds.UserID, &creds.Username,
			&creds.Password, &creds.Uri, &creds.Description); err != nil {
			return nil, err
		}
		result = append(result, creds)
	}

	return result, nil
}

func (s *sqliteStorage) DeleteCredential(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, _deleteCred, id)
	return err
}

func (s *sqliteStorage) CredentialData(ctx context.Context, id string) (*CredentialData, error) {
	creds := &CredentialData{}
	err := s.db.QueryRowContext(ctx, _selectCred, id).
		Scan(&creds.ID, &creds.UserID, &creds.Username, &creds.Password, &creds.Uri, &creds.Description)
	return creds, err
}

func (s *sqliteStorage) prepareDB(ctx context.Context, tables ...string) error {
	for _, t := range tables {
		if _, err := s.db.ExecContext(ctx, t); err != nil {
			return err
		}
	}
	return nil
}
