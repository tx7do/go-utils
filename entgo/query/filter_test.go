package entgo

import (
	"testing"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"

	"github.com/stretchr/testify/require"
)

func TestFilter(t *testing.T) {
	t.Run("MySQL_FilterEqual", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := filterEqual(s, "name", "tom")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`name` = ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "tom")
	})
	t.Run("PostgreSQL_FilterEqual", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := filterEqual(s, "name", "tom")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"name\" = $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "tom")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterNot", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := filterNot(s, "name", "tom")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE NOT (`users`.`name` = ?)", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "tom")
	})
	t.Run("PostgreSQL_FilterNot", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := filterNot(s, "name", "tom")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE NOT (\"users\".\"name\" = $1)", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "tom")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterIn", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := filterIn(s, "name", "[\"tom\", \"jimmy\", 123]")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`name` IN (?, ?, ?)", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "tom")
		require.Equal(t, args[1], "jimmy")
		require.Equal(t, args[2], float64(123))
	})
	t.Run("PostgreSQL_FilterIn", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := filterIn(s, "name", "[\"tom\", \"jimmy\", 123]")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"name\" IN ($1, $2, $3)", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "tom")
		require.Equal(t, args[1], "jimmy")
		require.Equal(t, args[2], float64(123))
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterNotIn", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := filterNotIn(s, "name", "[\"tom\", \"jimmy\", 123]")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`name` NOT IN (?, ?, ?)", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "tom")
		require.Equal(t, args[1], "jimmy")
		require.Equal(t, args[2], float64(123))
	})
	t.Run("PostgreSQL_FilterNotIn", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := filterNotIn(s, "name", "[\"tom\", \"jimmy\", 123]")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"name\" NOT IN ($1, $2, $3)", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "tom")
		require.Equal(t, args[1], "jimmy")
		require.Equal(t, args[2], float64(123))
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterGTE", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := filterGTE(s, "create_time", "2023-10-25")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`create_time` >= ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "2023-10-25")
	})
	t.Run("PostgreSQL_FilterGTE", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := filterGTE(s, "create_time", "2023-10-25")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"create_time\" >= $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "2023-10-25")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterGT", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := filterGT(s, "create_time", "2023-10-25")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`create_time` > ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "2023-10-25")
	})
	t.Run("PostgreSQL_FilterGT", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := filterGT(s, "create_time", "2023-10-25")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"create_time\" > $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "2023-10-25")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterLTE", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := filterLTE(s, "create_time", "2023-10-25")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`create_time` <= ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "2023-10-25")
	})
	t.Run("PostgreSQL_FilterLTE", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := filterLTE(s, "create_time", "2023-10-25")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"create_time\" <= $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "2023-10-25")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterLT", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := filterLT(s, "create_time", "2023-10-25")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`create_time` < ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "2023-10-25")
	})
	t.Run("PostgreSQL_FilterLT", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := filterLT(s, "create_time", "2023-10-25")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"create_time\" < $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "2023-10-25")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterRange", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := filterRange(s, "create_time", "[\"2023-10-25\", \"2024-10-25\"]")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`create_time` >= ? AND `users`.`create_time` <= ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "2023-10-25")
		require.Equal(t, args[1], "2024-10-25")
	})
	t.Run("PostgreSQL_FilterRange", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := filterRange(s, "create_time", "[\"2023-10-25\", \"2024-10-25\"]")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"create_time\" >= $1 AND \"users\".\"create_time\" <= $2", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "2023-10-25")
		require.Equal(t, args[1], "2024-10-25")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterIsNull", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := filterIsNull(s, "name", "true")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`name` IS NULL", query)
		require.Empty(t, args)
	})
	t.Run("PostgreSQL_FilterIsNull", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := filterIsNull(s, "name", "true")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"name\" IS NULL", query)
		require.Empty(t, args)
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterIsNotNull", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := filterIsNotNull(s, "name", "true")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE NOT (`users`.`name` IS NULL)", query)
		require.Empty(t, args)
	})
	t.Run("PostgreSQL_FilterIsNotNull", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := filterIsNotNull(s, "name", "true")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE NOT (\"users\".\"name\" IS NULL)", query)
		require.Empty(t, args)
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterContains", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := filterContains(s, "name", "L")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`name` LIKE ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "%L%")
	})
	t.Run("PostgreSQL_FilterContains", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := filterContains(s, "name", "L")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"name\" LIKE $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "%L%")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterInsensitiveContains", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := filterInsensitiveContains(s, "name", "L")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`name` COLLATE utf8mb4_general_ci LIKE ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "%l%")
	})
	t.Run("PostgreSQL_FilterInsensitiveContains", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := filterInsensitiveContains(s, "name", "L")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"name\" ILIKE $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "%l%")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterStartsWith", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := filterStartsWith(s, "name", "La")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`name` LIKE ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "La%")
	})
	t.Run("PostgreSQL_FilterStartsWith", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := filterStartsWith(s, "name", "La")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"name\" LIKE $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "La%")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterInsensitiveStartsWith", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := filterInsensitiveStartsWith(s, "name", "La")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`name` COLLATE utf8mb4_general_ci = ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "la%")
	})
	t.Run("PostgreSQL_FilterInsensitiveStartsWith", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := filterInsensitiveStartsWith(s, "name", "La")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"name\" ILIKE $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "la\\%")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterEndsWith", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := filterEndsWith(s, "name", "La")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`name` LIKE ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "%La")
	})
	t.Run("PostgreSQL_FilterEndsWith", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := filterEndsWith(s, "name", "La")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"name\" LIKE $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "%La")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterInsensitiveEndsWith", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := filterInsensitiveEndsWith(s, "name", "La")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`name` COLLATE utf8mb4_general_ci = ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "%la")
	})
	t.Run("PostgreSQL_FilterInsensitiveEndsWith", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := filterInsensitiveEndsWith(s, "name", "La")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"name\" ILIKE $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "\\%la")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterExact", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := filterExact(s, "name", "La")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`name` LIKE ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "La")
	})
	t.Run("PostgreSQL_FilterExact", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := filterExact(s, "name", "La")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"name\" LIKE $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "La")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterInsensitiveExact", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := filterInsensitiveExact(s, "name", "La")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`name` COLLATE utf8mb4_general_ci = ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "la")
	})
	t.Run("PostgreSQL_FilterInsensitiveExact", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := filterInsensitiveExact(s, "name", "La")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"name\" ILIKE $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "la")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterRegex", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := filterRegex(s, "name", "^(An?|The) +")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`name` REGEXP BINARY ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "^(An?|The) +")
	})
	t.Run("PostgreSQL_FilterRegex", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := filterRegex(s, "name", "^(An?|The) +")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"name\" ~ $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "^(An?|The) +")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterInsensitiveRegex", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := filterInsensitiveRegex(s, "name", "^(An?|The) +")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`name` REGEXP ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "^(an?|the) +")
	})
	t.Run("PostgreSQL_FilterInsensitiveRegex", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := filterInsensitiveRegex(s, "name", "^(An?|The) +")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"name\" ~* $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "^(an?|the) +")
	})
}
