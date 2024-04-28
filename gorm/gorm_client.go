package gorm

import (
	"fmt"

	"gorm.io/gorm"

	"gorm.io/plugin/opentelemetry/tracing"

	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
)

type Client struct {
	*gorm.DB

	err error
}

func NewClient(driverName, dsn string, enableMigrate, enableTrace, enableMetrics bool, gormCfg *gorm.Config) *Client {
	c := &Client{}

	if gormCfg == nil {
		gormCfg = &gorm.Config{}
	}

	c.err = c.createGormClient(driverName, dsn, enableMigrate, enableTrace, enableMetrics, gormCfg)

	return c
}

func (c *Client) Error() error {
	return c.err
}

// createGormClient 创建GORM的客户端
func (c *Client) createGormClient(driverName, dsn string, enableMigrate, enableTrace, enableMetrics bool, gormCfg *gorm.Config) error {
	var driver gorm.Dialector
	switch driverName {
	default:
		fallthrough
	case "mysql":
		driver = mysql.Open(dsn)
		break
	case "postgres":
		driver = postgres.Open(dsn)
		break
	case "clickhouse":
		driver = clickhouse.Open(dsn)
		break
	case "sqlite":
		driver = sqlite.Open(dsn)
		break
	case "sqlserver":
		driver = sqlserver.Open(dsn)
		break
	}

	client, err := gorm.Open(driver, gormCfg)
	if err != nil {
		return fmt.Errorf("failed opening connection to db: %v", err)
	}

	if enableTrace {
		var opts []tracing.Option
		if enableMetrics {
			opts = append(opts, tracing.WithoutMetrics())
		}

		if err = client.Use(tracing.NewPlugin(opts...)); err != nil {
			return fmt.Errorf("failed opening connection to db: %v", err)
		}
	}

	// 运行数据库迁移工具
	if enableMigrate {
		if err = client.AutoMigrate(
			getMigrateModels()...,
		); err != nil {
			return fmt.Errorf("failed creating schema resources: %v", err)
		}
	}

	c.DB = client

	return nil
}
