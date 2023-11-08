package entgo

import (
	"testing"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"

	"github.com/stretchr/testify/require"
)

func TestBuildSetNullUpdate(t *testing.T) {
	t.Run("MySQL_Set2", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Update("users")

		BuildSetNullUpdate(s, []string{"id", "username"})
		query, args := s.Query()
		require.Equal(t, "UPDATE `users` SET `id` = NULL, `username` = NULL", query)
		require.Empty(t, args)

	})
	t.Run("PostgreSQL_Set2", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Update("users")

		BuildSetNullUpdate(s, []string{"id", "username"})
		query, args := s.Query()
		require.Equal(t, `UPDATE "users" SET "id" = NULL, "username" = NULL`, query)
		require.Empty(t, args)
	})
}
