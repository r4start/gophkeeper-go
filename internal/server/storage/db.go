package storage

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	_ UserService = (*dbStorage)(nil)
	_ Storage     = (*dbStorage)(nil)

	_ Resource = (*dbResource)(nil)
)

const (
	_dbOperationTimeout = 250 * time.Millisecond

	_emptyOID = uint32(0)

	_addUser = `insert into users (id, login, key_salt, salt, secret) VALUES ('%s', '%s', '\x%s', '\x%s', '\x%s');`

	_getUserByLogin = `select id, login, salt, secret from users where is_deleted='false' and login=$1;`
	_getUserByID    = `select id, login, salt, secret from users where is_deleted='false' and id=$1;`

	_addNewResource = `insert into user_data (user_id, resource_id, data_id, salt) values('%s', '%s', '%d', '\x%s');`
	_getResource    = `select data_id, salt from user_data where resource_id=$1 and user_id=$2 and is_deleted='false';`
	_listResources  = `select resource_id, salt from user_data where user_id=$1 and is_deleted='false';`
	_deleteResource = `update user_data set is_deleted='true' where user_id=$1 and resource_id=$2;`
)

type dbStorage struct {
	dbConn *pgxpool.Pool
}

func NewDatabaseUserService(ctx context.Context, dsn string) (*dbStorage, error) {
	conn, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	return &dbStorage{
		dbConn: conn,
	}, nil
}

func (d *dbStorage) Add(ctx context.Context, login string, keySalt, salt, secret []byte) (*UserID, error) {
	c, cancel := context.WithTimeout(ctx, _dbOperationTimeout)
	defer cancel()

	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	tx, err := d.dbConn.Begin(c)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback(c)
	}()

	insertQuery := fmt.Sprintf(_addUser, id.String(), login,
		hex.EncodeToString(keySalt), hex.EncodeToString(salt), hex.EncodeToString(secret))
	if _, err := tx.Exec(c, insertQuery); err != nil {
		return nil, err
	}

	uid := UserID(id)
	return &uid, tx.Commit(c)
}

func (d *dbStorage) GetByLogin(ctx context.Context, login string) (*User, error) {
	c, cancel := context.WithTimeout(ctx, _dbOperationTimeout)
	defer cancel()

	var (
		user = &User{}
		id   uuid.UUID
	)
	row := d.dbConn.QueryRow(c, _getUserByLogin, login)
	if err := row.Scan(&id, &user.Login, &user.Salt, &user.Secret); err != nil {
		return nil, err
	}

	user.ID = UserID(id)

	return user, nil
}

func (d *dbStorage) GetByID(ctx context.Context, id string) (*User, error) {
	c, cancel := context.WithTimeout(ctx, _dbOperationTimeout)
	defer cancel()

	var (
		user   = &User{}
		userID uuid.UUID
	)
	row := d.dbConn.QueryRow(c, _getUserByID, id)
	if err := row.Scan(&userID, &user.Login, &user.Salt, &user.Secret); err != nil {
		return nil, err
	}

	user.ID = UserID(userID)

	return user, nil
}

func (d *dbStorage) Close() error {
	d.dbConn.Close()
	return nil
}

func (d *dbStorage) Create(ctx context.Context, user *UserID, salt []byte) (Resource, error) {
	resourceId, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	tx, err := d.dbConn.Begin(ctx)
	if err != nil {
		return nil, err
	}

	lo := tx.LargeObjects()
	oid, err := lo.Create(ctx, _emptyOID)
	if err != nil {
		if e := tx.Rollback(ctx); e != nil {
			err = multierror.Append(err, e)
		}
		return nil, err
	}

	addQuery := fmt.Sprintf(_addNewResource, user.String(), resourceId.String(), oid, hex.EncodeToString(salt))
	if _, err := tx.Exec(ctx, addQuery); err != nil {
		if e := tx.Rollback(ctx); e != nil {
			err = multierror.Append(err, e)
		}
		return nil, err
	}

	obj, err := lo.Open(ctx, oid, pgx.LargeObjectModeRead|pgx.LargeObjectModeWrite)
	if err != nil {
		if e := tx.Rollback(ctx); e != nil {
			err = multierror.Append(err, e)
		}
		return nil, err
	}

	return &dbResource{
		ctx:  ctx,
		tx:   tx,
		lo:   obj,
		id:   ResourceID(resourceId),
		salt: salt,
	}, nil
}

func (d *dbStorage) Open(ctx context.Context, user *UserID, id *ResourceID) (Resource, error) {
	tx, err := d.dbConn.Begin(ctx)
	if err != nil {
		return nil, err
	}

	var (
		oid  = _emptyOID
		salt []byte
	)
	if err := tx.QueryRow(ctx, _getResource, id, user).Scan(&oid, &salt); err != nil {
		if e := tx.Rollback(ctx); e != nil {
			err = multierror.Append(err, e)
		}
		return nil, err
	}

	lo := tx.LargeObjects()
	obj, err := lo.Open(ctx, oid, pgx.LargeObjectModeRead|pgx.LargeObjectModeWrite)
	if err != nil {
		return nil, err
	}

	return &dbResource{
		ctx:  ctx,
		tx:   tx,
		lo:   obj,
		id:   *id,
		salt: salt,
	}, nil
}

func (d *dbStorage) Delete(ctx context.Context, user *UserID, id *ResourceID) error {
	_, err := d.dbConn.Exec(ctx, _deleteResource, user, id)
	return err
}

func (d *dbStorage) List(ctx context.Context, userId *UserID) ([]Resource, error) {
	rows, err := d.dbConn.Query(ctx, _listResources, userId.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resources := make([]Resource, 0)

	for rows.Next() {
		var (
			id   uuid.UUID
			salt []byte
		)
		if err := rows.Scan(&id, &salt); err != nil {
			return nil, err
		}
		resources = append(resources, &dbResource{id: ResourceID(id), salt: salt})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return resources, nil
}

type dbResource struct {
	ctx       context.Context
	tx        pgx.Tx
	lo        *pgx.LargeObject
	id        ResourceID
	salt      []byte
	isDeleted bool
}

func (d *dbResource) Close() error {
	return d.tx.Commit(d.ctx)
}

func (d *dbResource) Write(p []byte) (n int, err error) {
	return d.lo.Write(p)
}

func (d *dbResource) Read(p []byte) (n int, err error) {
	return d.lo.Read(p)
}

func (d *dbResource) GetId() *ResourceID {
	return &d.id
}

func (d *dbResource) IsDeleted() bool {
	return d.isDeleted
}

func (d *dbResource) Salt() ([]byte, error) {
	return d.salt, nil
}
