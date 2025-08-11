package postgres

import (
	"testing"
	"time"

	"github.com/rakunlabs/tummy"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/worldline-go/saz/internal/config"
	"github.com/worldline-go/saz/internal/service"
	"github.com/worldline-go/test/container/containerpostgres"
	"github.com/worldline-go/test/utils/dbutils"
)

type PostgresSuite struct {
	suite.Suite
	container *containerpostgres.Container
}

func (s *PostgresSuite) SetupSuite() {
	tummy.Enable()
	s.container = containerpostgres.New(s.T())
	s.container.ExecuteFolder(s.T(), "./migrations", dbutils.WithValues(map[string]string{
		"table_prefix": "",
	}))
}

func TestExampleTestSuitePostgres(t *testing.T) {
	suite.Run(t, new(PostgresSuite))
}

func (s *PostgresSuite) TearDownSuite() {
	s.container.Stop(s.T())
}

func (s *PostgresSuite) Test_Save() {
	postgres, err := conn(&config.StorePostgres{}, s.container.Sqlx())
	require.NoError(s.T(), err)
	require.NotNil(s.T(), postgres)

	// Test saving a note
	note := &service.Note{
		ID:      "test-note",
		Name:    "Test Note",
		Content: service.Content{Cells: []service.Cell{{ID: "cell1", DBType: "postgres", Content: "SELECT * FROM test"}}},
		Path:    "test-note",
	}

	tummy.Pause()
	tummy.SetTime(tummy.Now().Truncate(time.Microsecond))
	now := tummy.Now()

	err = postgres.Save(service.ContextWithUser(s.T().Context(), "test-user"), note)
	require.NoError(s.T(), err)

	// Verify the note was saved
	savedNote, err := postgres.Get(s.T().Context(), note.ID)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), savedNote)
	require.Equal(s.T(), note.ID, savedNote.ID)
	require.Equal(s.T(), note.Name, savedNote.Name)
	require.Equal(s.T(), note.Content, savedNote.Content)
	require.Equal(s.T(), note.Path, savedNote.Path)
	require.Equal(s.T(), "test-user", savedNote.UpdatedBy.V)

	getCreatedAt := savedNote.CreatedAt.V
	require.Equal(s.T(), getCreatedAt.IsZero(), false, "CreatedAt should not be zero")

	getUpdatedAt := savedNote.UpdatedAt.V
	require.Equal(s.T(), getUpdatedAt.IsZero(), false, "UpdatedAt should not be zero")

	tummy.AddDuration(10 * time.Second)

	// Change the content
	note.Content = service.Content{Cells: []service.Cell{{ID: "celo", DBType: "postgres", Content: "SELECT * FROM test"}}}
	err = postgres.Save(service.ContextWithUser(s.T().Context(), "test-user-2"), note)
	require.NoError(s.T(), err)

	// Get with path
	notePath := "test-note"
	noteByPath, err := postgres.GetWithPath(s.T().Context(), notePath)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), noteByPath)
	require.Equal(s.T(), note.ID, noteByPath.ID)
	require.Equal(s.T(), note.Name, noteByPath.Name)
	require.Equal(s.T(), note.Content, noteByPath.Content)
	require.Equal(s.T(), note.Path, noteByPath.Path)
	require.Equal(s.T(), "test-user-2", noteByPath.UpdatedBy.V)

	// CreatedAt should
	require.Equal(s.T(), now.Truncate(time.Microsecond), noteByPath.CreatedAt.V.Time, "CreatedAt should be the same with now")
	require.Equal(s.T(), getCreatedAt, noteByPath.CreatedAt.V, "CreatedAt should be the same older get")

	// UpdatedAt should be 10 seconds different
	require.Equal(s.T(), noteByPath.UpdatedAt.V.Sub(getUpdatedAt.Time), 10*time.Second, "UpdatedAt should be different")
}
