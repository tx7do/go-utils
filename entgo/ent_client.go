package entgo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	entSql "entgo.io/ent/dialect/sql"
)

type entClientInterface interface {
	Close() error
}

type EntClient[T entClientInterface] struct {
	db  T
	drv *entSql.Driver
}

func NewEntClient[T entClientInterface](db T, drv *entSql.Driver) *EntClient[T] {
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

func (c *EntClient[T]) Close() error {
	return c.db.Close()
}

func (c *EntClient[T]) Query(ctx context.Context, query string, args, v any) error {
	return c.Driver().Query(ctx, query, args, v)
}

func (c *EntClient[T]) Exec(ctx context.Context, query string, args, v any) error {
	return c.Driver().Exec(ctx, query, args, v)
}

// CreateDriver 创建数据库驱动
func CreateDriver(driverName, dsn string, maxIdleConnections, maxOpenConnections int, connMaxLifetime time.Duration) (*entSql.Driver, error) {
	drv, err := entSql.Open(driverName, dsn)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed opening connection to db: %v", err))
	}

	db := drv.DB()
	// 连接池中最多保留的空闲连接数量
	db.SetMaxIdleConns(maxIdleConnections)
	// 连接池在同一时间打开连接的最大数量
	db.SetMaxOpenConns(maxOpenConnections)
	// 连接可重用的最大时间长度
	db.SetConnMaxLifetime(connMaxLifetime)

	return drv, nil
}
