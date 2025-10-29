package entgo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"

	"github.com/XSAM/otelsql"

	entSql "entgo.io/ent/dialect/sql"
)

type EntClientInterface interface {
	Close() error
}

type EntClient[T EntClientInterface] struct {
	db  T
	drv *entSql.Driver
}

func NewEntClient[T EntClientInterface](db T, drv *entSql.Driver) *EntClient[T] {
	return &EntClient[T]{
		db:  db,
		drv: drv,
	}
}

func (c *EntClient[T]) Client() T {
	return c.db
}

func (c *EntClient[T]) Driver() *entSql.Driver {
	return c.drv
}

func (c *EntClient[T]) DB() *sql.DB {
	return c.drv.DB()
}

// Close 关闭数据库连接
func (c *EntClient[T]) Close() error {
	return c.db.Close()
}

// Query 查询数据
func (c *EntClient[T]) Query(ctx context.Context, query string, args, v any) error {
	return c.Driver().Query(ctx, query, args, v)
}

func (c *EntClient[T]) Exec(ctx context.Context, query string, args, v any) error {
	return c.Driver().Exec(ctx, query, args, v)
}

// SetConnectionOption 设置连接配置
func (c *EntClient[T]) SetConnectionOption(maxIdleConnections, maxOpenConnections int, connMaxLifetime time.Duration) {
	// 连接池中最多保留的空闲连接数量
	c.DB().SetMaxIdleConns(maxIdleConnections)
	// 连接池在同一时间打开连接的最大数量
	c.DB().SetMaxOpenConns(maxOpenConnections)
	// 连接可重用的最大时间长度
	c.DB().SetConnMaxLifetime(connMaxLifetime)
}

func driverNameToSemConvKeyValue(driverName string) attribute.KeyValue {
	switch driverName {
	case "mariadb":
		return semconv.DBSystemMariaDB
	case "mysql":
		return semconv.DBSystemMySQL
	case "postgresql":
		return semconv.DBSystemPostgreSQL
	case "sqlite":
		return semconv.DBSystemSqlite
	default:
		return semconv.DBSystemKey.String(driverName)
	}
}

// CreateDriver 创建数据库驱动
func CreateDriver(driverName, dsn string, enableTrace, enableMetrics bool) (*entSql.Driver, error) {
	var db *sql.DB
	var drv *entSql.Driver
	var err error

	if enableTrace {
		// Connect to database
		if db, err = otelsql.Open(driverName, dsn, otelsql.WithAttributes(
			driverNameToSemConvKeyValue(driverName),
		)); err != nil {
			return nil, errors.New(fmt.Sprintf("failed opening connection to db: %v", err))
		}

		drv = entSql.OpenDB(driverName, db)
	} else {
		if drv, err = entSql.Open(driverName, dsn); err != nil {
			return nil, errors.New(fmt.Sprintf("failed opening connection to db: %v", err))
		}

		db = drv.DB()
	}

	// Register DB stats to meter
	if enableMetrics {
		err = otelsql.RegisterDBStatsMetrics(db, otelsql.WithAttributes(
			driverNameToSemConvKeyValue(driverName),
		))
		if err != nil {
			return nil, errors.New(fmt.Sprintf("failed register otel meter: %v", err))
		}
	}

	return drv, nil
}

type Rollbacker interface {
	Rollback() error
}

// Rollback calls to tx.Rollback and wraps the given error
func Rollback[T Rollbacker](tx T, err error) error {
	if rerr := tx.Rollback(); rerr != nil {
		if err == nil {
			err = rerr
		} else {
			err = fmt.Errorf("%w: rollback failed: %v", err, rerr)
		}
	}
	return err
}

// QueryAllChildrenIds 使用CTE递归查询所有子节点ID
func QueryAllChildrenIds[T EntClientInterface](ctx context.Context, entClient *EntClient[T], tableName string, parentID uint32) ([]uint32, error) {
	var query string
	switch entClient.Driver().Dialect() {
	case dialect.MySQL:
		query = fmt.Sprintf(`
			WITH RECURSIVE all_descendants AS (
				SELECT 
					id,
					parent_id,
					name,
					1 AS depth
				FROM %s
				WHERE parent_id = ?
				
				UNION ALL
				
				SELECT 
					p.id,
					p.parent_id,
					p.name,
					ad.depth + 1 AS depth
				FROM %s p
					INNER JOIN all_descendants ad
				ON p.parent_id = ad.id
			)
			SELECT id FROM all_descendants;
		`, tableName, tableName)

	case dialect.Postgres:
		query = fmt.Sprintf(`
        WITH RECURSIVE all_descendants AS (
            SELECT *
			FROM %s
			WHERE parent_id = $1
            UNION ALL
            SELECT p.*
			FROM %s p
            	INNER JOIN all_descendants ad
			ON p.parent_id = ad.id
        )
        SELECT id FROM all_descendants;
    `, tableName, tableName)
	}

	rows := &sql.Rows{}
	if err := entClient.Query(ctx, query, []any{parentID}, rows); err != nil {
		log.Errorf("query child nodes failed: %s", err.Error())
		return nil, errors.New("query child nodes failed: " + err.Error())
	}
	defer rows.Close()

	childIDs := make([]uint32, 0)
	for rows.Next() {
		var id uint32

		if err := rows.Scan(&id); err != nil {
			log.Errorf("scan child node failed: %s", err.Error())
			return nil, errors.New("scan child node failed")
		}

		childIDs = append(childIDs, id)
	}

	return childIDs, nil
}
