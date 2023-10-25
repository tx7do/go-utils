package entgo

import (
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
)

type entClientInterface interface {
	Close() error
}

type EntClient[T entClientInterface] struct {
	db  T
	drv *sql.Driver
}

func NewEntClient[T entClientInterface](db T, drv *sql.Driver) *EntClient[T] {
	return &EntClient[T]{
		db:  db,
		drv: drv,
	}
}

func (c *EntClient[T]) Client() T {
	return c.db
}

func (c *EntClient[T]) Driver() *sql.Driver {
	return c.drv
}

func (c *EntClient[T]) Close() {
	_ = c.db.Close()
}

// CreateDriver 创建数据库驱动
func CreateDriver(driverName, dsn string, maxIdleConnections, maxOpenConnections int, connMaxLifetime time.Duration) (*sql.Driver, error) {
	drv, err := sql.Open(driverName, dsn)
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
