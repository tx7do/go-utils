package entgo

import (
	"fmt"
	"testing"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"

	"github.com/stretchr/testify/require"
)

func TestFilter(t *testing.T) {
	t.Run("MySQL_FilterEqual", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterEqual(s, p, "name", "tom")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`name` = ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "tom")
	})
	t.Run("PostgreSQL_FilterEqual", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterEqual(s, p, "name", "tom")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"name\" = $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "tom")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterNot", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterNot(s, p, "name", "tom")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE NOT `users`.`name` = ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "tom")
	})
	t.Run("PostgreSQL_FilterNot", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterNot(s, p, "name", "tom")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE NOT \"users\".\"name\" = $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "tom")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterIn", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterIn(s, p, "name", "[\"tom\", \"jimmy\", 123]")
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

		p := sql.P()

		p = filterIn(s, p, "name", "[\"tom\", \"jimmy\", 123]")
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

		p := sql.P()

		p = filterNotIn(s, p, "name", "[\"tom\", \"jimmy\", 123]")
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

		p := sql.P()

		p = filterNotIn(s, p, "name", "[\"tom\", \"jimmy\", 123]")
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

		p := sql.P()

		p = filterGTE(s, p, "create_time", "2023-10-25")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`create_time` >= ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "2023-10-25")
	})
	t.Run("PostgreSQL_FilterGTE", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterGTE(s, p, "create_time", "2023-10-25")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"create_time\" >= $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "2023-10-25")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterGT", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterGT(s, p, "create_time", "2023-10-25")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`create_time` > ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "2023-10-25")
	})
	t.Run("PostgreSQL_FilterGT", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterGT(s, p, "create_time", "2023-10-25")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"create_time\" > $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "2023-10-25")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterLTE", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterLTE(s, p, "create_time", "2023-10-25")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`create_time` <= ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "2023-10-25")
	})
	t.Run("PostgreSQL_FilterLTE", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterLTE(s, p, "create_time", "2023-10-25")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"create_time\" <= $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "2023-10-25")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterLT", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterLT(s, p, "create_time", "2023-10-25")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`create_time` < ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "2023-10-25")
	})
	t.Run("PostgreSQL_FilterLT", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterLT(s, p, "create_time", "2023-10-25")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"create_time\" < $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "2023-10-25")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterRange", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterRange(s, p, "create_time", "[\"2023-10-25\", \"2024-10-25\"]")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`create_time` >= ? AND `users`.`create_time` <= ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "2023-10-25")
		require.Equal(t, args[1], "2024-10-25")
	})
	t.Run("PostgreSQL_FilterRange", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterRange(s, p, "create_time", "[\"2023-10-25\", \"2024-10-25\"]")
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

		p := sql.P()

		p = filterIsNull(s, p, "name", "true")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`name` IS NULL", query)
		require.Empty(t, args)
	})
	t.Run("PostgreSQL_FilterIsNull", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterIsNull(s, p, "name", "true")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"name\" IS NULL", query)
		require.Empty(t, args)
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterIsNotNull", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterIsNotNull(s, p, "name", "true")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE NOT `users`.`name` IS NULL", query)
		require.Empty(t, args)
	})
	t.Run("PostgreSQL_FilterIsNotNull", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterIsNotNull(s, p, "name", "true")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE NOT \"users\".\"name\" IS NULL", query)
		require.Empty(t, args)
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterContains", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterContains(s, p, "name", "L")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`name` LIKE ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "%L%")
	})
	t.Run("PostgreSQL_FilterContains", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterContains(s, p, "name", "L")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"name\" LIKE $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "%L%")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterInsensitiveContains", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterInsensitiveContains(s, p, "name", "L")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`name` COLLATE utf8mb4_general_ci LIKE ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "%l%")
	})
	t.Run("PostgreSQL_FilterInsensitiveContains", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterInsensitiveContains(s, p, "name", "L")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"name\" ILIKE $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "%l%")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterStartsWith", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterStartsWith(s, p, "name", "La")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`name` LIKE ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "La%")
	})
	t.Run("PostgreSQL_FilterStartsWith", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterStartsWith(s, p, "name", "La")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"name\" LIKE $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "La%")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterInsensitiveStartsWith", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterInsensitiveStartsWith(s, p, "name", "La")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`name` COLLATE utf8mb4_general_ci = ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "la%")
	})
	t.Run("PostgreSQL_FilterInsensitiveStartsWith", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterInsensitiveStartsWith(s, p, "name", "La")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"name\" ILIKE $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "la\\%")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterEndsWith", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterEndsWith(s, p, "name", "La")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`name` LIKE ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "%La")
	})
	t.Run("PostgreSQL_FilterEndsWith", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterEndsWith(s, p, "name", "La")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"name\" LIKE $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "%La")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterInsensitiveEndsWith", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterInsensitiveEndsWith(s, p, "name", "La")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`name` COLLATE utf8mb4_general_ci = ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "%la")
	})
	t.Run("PostgreSQL_FilterInsensitiveEndsWith", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterInsensitiveEndsWith(s, p, "name", "La")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"name\" ILIKE $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "\\%la")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterExact", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterExact(s, p, "name", "La")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`name` LIKE ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "La")
	})
	t.Run("PostgreSQL_FilterExact", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterExact(s, p, "name", "La")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"name\" LIKE $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "La")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterInsensitiveExact", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterInsensitiveExact(s, p, "name", "La")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`name` COLLATE utf8mb4_general_ci = ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "la")
	})
	t.Run("PostgreSQL_FilterInsensitiveExact", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterInsensitiveExact(s, p, "name", "La")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"name\" ILIKE $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "la")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterRegex", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterRegex(s, p, "name", "^(An?|The) +")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`name` REGEXP BINARY ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "^(An?|The) +")
	})
	t.Run("PostgreSQL_FilterRegex", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterRegex(s, p, "name", "^(An?|The) +")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"name\" ~ $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "^(An?|The) +")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterInsensitiveRegex", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterInsensitiveRegex(s, p, "name", "^(An?|The) +")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `users` WHERE `users`.`name` REGEXP ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "^(an?|the) +")
	})
	t.Run("PostgreSQL_FilterInsensitiveRegex", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("users"))

		p := sql.P()

		p = filterInsensitiveRegex(s, p, "name", "^(An?|The) +")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"users\" WHERE \"users\".\"name\" ~* $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "^(an?|the) +")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterDatePart", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("publishes"))

		p := sql.P()

		p = filterDatePart(s, p, "date", "pub_date")
		p.EQ("", "2023-01-01")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `publishes` WHERE DATE(`publishes`.`pub_date`) = ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "2023-01-01")
	})
	t.Run("PostgreSQL_FilterDatePart", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("publishes"))

		p := sql.P()

		p = filterDatePart(s, p, "date", "pub_date")
		p.EQ("", "2023-01-01")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"publishes\" WHERE EXTRACT('DATE' FROM \"publishes\".\"pub_date\") = $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "2023-01-01")
	})

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////

	t.Run("MySQL_FilterJsonb", func(t *testing.T) {
		s := sql.Dialect(dialect.MySQL).Select("*").From(sql.Table("app_profile"))

		p := sql.P()

		p = filterJsonb(s, p, "daily_email", "preferences")
		p.EQ("", "true")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM `app_profile` WHERE JSON_EXTRACT(`app_profile`.`preferences`, '$.daily_email') = ?", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "true")
	})
	t.Run("PostgreSQL_FilterJsonb", func(t *testing.T) {
		s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("app_profile"))

		p := sql.P()

		p = filterJsonb(s, p, "daily_email", "preferences")
		p.EQ("", "true")
		s.Where(p)

		query, args := s.Query()
		require.Equal(t, "SELECT * FROM \"app_profile\" WHERE \"app_profile\".\"preferences\" -> daily_email = $1", query)
		require.NotEmpty(t, args)
		require.Equal(t, args[0], "true")
	})
}

func TestFilterJsonbField(t *testing.T) {
	s := sql.Dialect(dialect.Postgres).Select("*").From(sql.Table("app_profile"))
	str := filterJsonbField(s, "daily_email", "preferences")
	fmt.Println(str)
}
