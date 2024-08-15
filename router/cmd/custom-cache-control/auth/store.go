package auth

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type dbUser struct {
	Id           int64          `db:"id"`
	UserLogin    sql.NullString `db:"userLogin"`
	CustId       sql.NullInt64  `db:"custId"`
	AvatarUrl    sql.NullString `db:"avatarUrl"`
	Extra        sql.NullString `db:"extra"`
	EncryptedPid sql.NullString `db:"encryptedPid"`
}

type saveDbUser struct {
	UserLogin    sql.NullString
	CustId       sql.NullInt64
	EncryptedPid sql.NullString
}

type store struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func newStore(logger *zap.Logger) (*store, error) {
	dbHost, dbPort, dbUser, dbPassword, dbDatabase := "localhost", 3306, "root", "", "cbs"

	if value, exists := os.LookupEnv("DB_HOST"); exists {
		dbHost = value
	}

	if value, exists := os.LookupEnv("DB_PORT"); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			dbPort = intValue
		}
	}

	if value, exists := os.LookupEnv("DB_USER"); exists {
		dbUser = value
	}

	if value, exists := os.LookupEnv("DB_PASSWORD"); exists {
		dbPassword = value
	}

	if value, exists := os.LookupEnv("DB_DATABASE"); exists {
		dbDatabase = value
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", dbUser, dbPassword, dbHost, dbPort, dbDatabase)

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("could not connect to db: %v", err)
	}
	return &store{db, logger}, nil
}

func (s *store) find(ctx context.Context, userLogin string) (*dbUser, error) {
	s.logger.Debug("finding db user", zap.String("userLogin", userLogin))
	var result dbUser

	err := sqlx.GetContext(ctx, s.db, &result, "SELECT id, userLogin, custId, avatarUrl, extra, encryptedPid FROM User WHERE userLogin = ?;", userLogin)
	if err != nil {
		return nil, fmt.Errorf("could not get db user: %w", err)
	}

	return &result, err
}

func (s *store) save(ctx context.Context, user saveDbUser) (*dbUser, error) {
	s.logger.Debug("saving db user", zap.String("userLogin", user.UserLogin.String))

	result, err := s.db.ExecContext(ctx, "INSERT INTO User(userLogin, custId, encryptedPid) VALUES (?,?,?);", user.UserLogin, user.CustId, user.EncryptedPid)

	if err != nil {
		return nil, fmt.Errorf("error while saving user: %v", err)
	}

	lastInsertedId, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("error while getting last inserted user id: %v", err)
	}

	return &dbUser{
		Id:           lastInsertedId,
		UserLogin:    user.UserLogin,
		CustId:       user.CustId,
		EncryptedPid: user.EncryptedPid,
		AvatarUrl:    sql.NullString{String: "", Valid: false},
		Extra:        sql.NullString{String: "", Valid: false},
	}, nil
}
