package database

import (
	"strconv"
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/worldline-go/saz/internal/service"
	"github.com/worldline-go/test/container/containerpostgres"
)

type DatabaseSuite struct {
	suite.Suite
	container *containerpostgres.Container
	Database  Database
}

func (s *DatabaseSuite) SetupSuite() {
	s.container = containerpostgres.New(s.T())
	s.container.ExecuteFolder(s.T(), "testdata")

	s.Database = Database{
		DB: map[string]*DatabaseInfo{
			"postgres": {
				DB:          s.container.Sqlx(),
				PlaceHolder: PlaceHolder(s.container.Sqlx().DriverName()),
			},
		},
	}
}

func TestDatabase(t *testing.T) {
	suite.Run(t, new(DatabaseSuite))
}

func (s *DatabaseSuite) TearDownSuite() {
	s.container.Stop(s.T())
}

func (s *DatabaseSuite) TearDownTest() {
	_, err := s.container.Sqlx().ExecContext(s.T().Context(), "TRUNCATE TABLE events")
	require.NoError(s.T(), err)
	_, err = s.container.Sqlx().ExecContext(s.T().Context(), "TRUNCATE TABLE events_copy")
	require.NoError(s.T(), err)
}

func (s *DatabaseSuite) TestCopyEventsEqualCounts() {
	// add data in the events
	batch := QueryBuilder("events", []string{"id", "name", "created_at"}, s.Database.DB["postgres"].PlaceHolder)

	n := 4
	batchQuery := batch(n)
	var args []any
	for i := range n {
		args = append(args,
			ulid.Make().String(),
			"test_event_"+strconv.Itoa(i),
			"2024-01-01 00:00:00Z",
		)
	}

	_, err := s.container.Sqlx().ExecContext(s.T().Context(), batchQuery, args...)
	require.NoError(s.T(), err)

	columns, rows, err := s.Database.IterGet(s.T().Context(), "postgres", "select * from events", service.MapType{})
	require.NoError(s.T(), err, "iterGet failed")

	result, err := s.Database.IterSet(s.T().Context(), "postgres", "events_copy", true, service.SkipError{}, service.MapType{}, 2, columns, rows)
	require.NoError(s.T(), err, "iterSet failed")
	require.NotNil(s.T(), result)

	require.Equal(s.T(), int64(4), result.RowsAffected())
}

func (s *DatabaseSuite) TestCopyEventsDiffCounts() {
	// add data in the events
	batch := QueryBuilder("events", []string{"id", "name", "created_at"}, s.Database.DB["postgres"].PlaceHolder)

	n := 11
	batchQuery := batch(n)
	var args []any
	for i := range n {
		args = append(args,
			ulid.Make().String(),
			"test_event_"+strconv.Itoa(i),
			"2024-01-01 00:00:00Z",
		)
	}

	_, err := s.container.Sqlx().ExecContext(s.T().Context(), batchQuery, args...)
	require.NoError(s.T(), err)

	columns, rows, err := s.Database.IterGet(s.T().Context(), "postgres", "select * from events", service.MapType{})
	require.NoError(s.T(), err, "iterGet failed")

	result, err := s.Database.IterSet(s.T().Context(), "postgres", "events_copy", true, service.SkipError{}, service.MapType{}, 3, columns, rows)
	require.NoError(s.T(), err, "iterSet failed")
	require.NotNil(s.T(), result)

	require.Equal(s.T(), int64(11), result.RowsAffected())
}
