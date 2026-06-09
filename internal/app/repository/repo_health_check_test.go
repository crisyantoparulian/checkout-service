package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type HealthCheckTestSuite struct {
	suite.Suite
	db   *sqlx.DB
	mock sqlmock.Sqlmock
	repo HealthCheck
}

func (s *HealthCheckTestSuite) SetupTest() {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	assert.NoError(s.T(), err)

	s.db = sqlx.NewDb(db, "postgres")
	s.mock = mock
	s.repo = NewHealthCheckRepository(s.db)
}

func (s *HealthCheckTestSuite) TearDownTest() {
	_ = s.db.Close()
}

func (s *HealthCheckTestSuite) TestPingDB_Success() {
	s.mock.ExpectPing()

	err := s.repo.PingDB(context.Background())

	assert.NoError(s.T(), err)
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func TestHealthCheckRepoTestSuite(t *testing.T) {
	suite.Run(t, new(HealthCheckTestSuite))
}

func TestPanic_NewHealthCheckRepository(t *testing.T) {
	assert.Panics(t, func() {
		NewHealthCheckRepository(nil)
	})
}
