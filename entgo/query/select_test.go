package entgo

import (
	"testing"

	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
)

func TestBuildFieldSelect(t *testing.T) {
	t.Run("MySQL_2Fields", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		BuildFieldSelect(s, []string{"id", "username"})
		query, args := s.Query()
		require.Equal(t, "SELECT `id`, `username` FROM `users`", query)
		require.Empty(t, args)

	})
	t.Run("PostgreSQL_2Fields", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		BuildFieldSelect(s, []string{"id", "username"})
		query, args := s.Query()
		require.Equal(t, `SELECT "id", "username" FROM "users"`, query)
		require.Empty(t, args)
	})

	t.Run("MySQL_AllFields", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		BuildFieldSelect(s, []string{})
		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users`", query)
		require.Empty(t, args)

	})
	t.Run("PostgreSQL_AllFields", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		BuildFieldSelect(s, []string{})
		query, args := s.Query()
		require.Equal(t, `SELECT * FROM "users"`, query)
		require.Empty(t, args)
	})
}
