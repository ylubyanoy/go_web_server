package postgres

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ylubyanoy/go_web_server/internal/data"
	"go.uber.org/zap"
)

var (
	// ErrRecordNotFound ...
	ErrRecordNotFound = errors.New("record not found")
)

// PostgresRepository ...
type PostgresRepository struct {
	conn   *pgxpool.Pool
	logger *zap.SugaredLogger
}

// NewPostgresRepository is create new connect to DB
func NewPostgresRepository(dbURL string, logger *zap.SugaredLogger) (*PostgresRepository, error) {
	conn, err := pgxpool.Connect(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	s := &PostgresRepository{
		conn:   conn,
		logger: logger,
	}

	return s, nil
}

// Close is closing connections to DB
func (repo *PostgresRepository) Close() {
	if repo.conn != nil {
		repo.conn.Close()
	}
}

// Create inserts the given user into the database
func (repo *PostgresRepository) Create(ctx context.Context, user *data.User) error {
	uuidValue, _ := uuid.NewV4()
	user.ID = uuidValue.String()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	repo.logger.Info("creating user ", user)
	query := "insert into users (id, email, username, password, tokenhash, createdat, updatedat) values ($1, $2, $3, $4, $5, $6, $7) returning id"
	return repo.conn.QueryRow(ctx, query, user.ID, user.Email, user.Username, user.Password, user.TokenHash, user.CreatedAt, user.UpdatedAt).Scan(&user.ID)
}

// GetUserByEmail retrieves the user object having the given email, else returns error
func (repo *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*data.User, error) {
	repo.logger.Debug("querying for user with email ", email)
	query := "select id, email, username, password, isverified from users where email = $1"
	var user data.User
	if err := repo.conn.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.Username, &user.Password, &user.IsVerified); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	repo.logger.Debug("read users ", user)
	return &user, nil
}

// GetUserByID retrieves the user object having the given ID, else returns error
func (repo *PostgresRepository) GetUserByID(ctx context.Context, userID string) (*data.User, error) {
	repo.logger.Debug("querying for user with id ", userID)
	query := "select  id, email, username from users where id = $1"
	var user data.User
	if err := repo.conn.QueryRow(ctx, query, userID).Scan(&user.ID, &user.Email, &user.Username); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	repo.logger.Debug("read users ", user)
	return &user, nil
}

// UpdateUsername updates the username of the given user
func (repo *PostgresRepository) UpdateUsername(ctx context.Context, user *data.User) error {
	repo.logger.Debug("updating user with id ", user.ID)
	user.UpdatedAt = time.Now()
	query := "update users set username = $1, updatedat = $2 where id = $3"
	if err := repo.conn.QueryRow(ctx, query, user.Username, user.UpdatedAt, user.ID).Scan(&user.ID); err != nil {
		if err == pgx.ErrNoRows {
			return ErrRecordNotFound
		}
		return err
	}
	repo.logger.Debug("updated user ", user)
	return nil
}

// UpdateUserVerificationStatus updates user verification status to true
func (repo *PostgresRepository) UpdateUserVerificationStatus(ctx context.Context, email string, status bool) error {
	repo.logger.Debug("updating verification status user with email ", email)
	var user data.User
	query := "update users set isverified = $1 where email = $2 returning id"
	if err := repo.conn.QueryRow(ctx, query, status, email).Scan(&user.ID); err != nil {
		if err == pgx.ErrNoRows {
			return ErrRecordNotFound
		}
		return err
	}
	repo.logger.Debug("updated user ", user)
	return nil
}

// StoreMailVerificationData adds a verification data to db
func (repo *PostgresRepository) StoreVerificationData(ctx context.Context, verificationData *data.VerificationData) error {
	var vData data.VerificationData
	query := "insert into verifications(email, code, expiresat, type) values($1, $2, $3, $4) returning email"
	return repo.conn.QueryRow(ctx, query, verificationData.Email, verificationData.Code, verificationData.ExpiresAt, strconv.Itoa(int(verificationData.Type))).Scan(&vData.Email)
}

// GetMailVerificationCode retrieves the stored verification code.
func (repo *PostgresRepository) GetVerificationData(ctx context.Context, email string, verificationDataType data.VerificationDataType) (*data.VerificationData, error) {
	query := "select * from verifications where email = $1 and type = $2"
	var verificationData data.VerificationData
	if err := repo.conn.QueryRow(ctx, query, email, strconv.Itoa(int(verificationDataType))).Scan(&verificationData.Email, &verificationData.Code, &verificationData.ExpiresAt, &verificationData.Type); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &verificationData, nil
}

// DeleteMailVerificationData deletes a used verification data
func (repo *PostgresRepository) DeleteVerificationData(ctx context.Context, email string, verificationDataType data.VerificationDataType) error {
	query := "delete from verifications where email = $1 and type = $2 returning email"
	var verificationData data.VerificationData
	return repo.conn.QueryRow(ctx, query, email, verificationDataType).Scan(&verificationData.Email)
}

// UpdatePassword updates the user password
func (repo *PostgresRepository) UpdatePassword(ctx context.Context, userID string, password string, tokenHash string) error {
	repo.logger.Debug("updating password for user ", userID)
	query := "update users set password = $1, tokenhash = $2 where id = $3 returning id"
	var user data.User
	if err := repo.conn.QueryRow(ctx, query, password, tokenHash, userID).Scan(&user.ID); err != nil {
		if err == pgx.ErrNoRows {
			return ErrRecordNotFound
		}
		return err
	}
	repo.logger.Debug("updated user", user)
	return nil
}
