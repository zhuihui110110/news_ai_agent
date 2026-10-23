// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: MySQL 工厂(参照 stock_ai factory.GetMySQLFactoryOr 单例模式,
// 集中持有 *gorm.DB 并提供各表 Repository 访问入口)

package factory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"news_ai_server/internal/domain/infrastructure/mysql/model"
	"news_ai_server/internal/domain/infrastructure/mysql/repository"
	"news_ai_server/internal/innerpkg/options"
)

var (
	mysqlFactory *datastore
	mysqlOnce    sync.Once
)

// Factory 仓储工厂接口(一个 accessor 对应一张表)
type Factory interface {
	NewsTaskRepository() repository.NewsTaskRepository
	NewsSourceArticleRepository() repository.NewsSourceArticleRepository
	NewsArticleRepository() repository.NewsArticleRepository
	NewsDraftRepository() repository.NewsDraftRepository
	Close() error
}

// datastore 工厂实现
type datastore struct {
	db *gorm.DB

	newsTaskRepository         repository.NewsTaskRepository
	newsSourceArticleRepository repository.NewsSourceArticleRepository
	newsArticleRepository      repository.NewsArticleRepository
	newsDraftRepository        repository.NewsDraftRepository
}

// GetMySQLFactoryOr 获取/初始化 MySQL 工厂单例
func GetMySQLFactoryOr(opts *options.MySQLOptions) (Factory, error) {
	if opts == nil {
		return nil, fmt.Errorf("mysql options 为空")
	}
	var err error
	mysqlOnce.Do(func() {
		db, e := gorm.Open(mysql.Open(opts.DSN()), &gorm.Config{
			Logger: logger.Default.LogMode(logger.LogLevel(opts.LogLevel)),
		})
		if e != nil {
			err = fmt.Errorf("连接 MySQL 失败: %w", e)
			return
		}
		sqlDB, e := db.DB()
		if e != nil {
			err = e
			return
		}
		sqlDB.SetMaxIdleConns(opts.MaxIdleConn)
		sqlDB.SetMaxOpenConns(opts.MaxOpenConn)
		sqlDB.SetConnMaxLifetime(24 * time.Hour)

		if opts.Migrate {
			if e = autoMigrate(db); e != nil {
				err = fmt.Errorf("自动建表失败: %w", e)
				return
			}
		}
		mysqlFactory = &datastore{db: db}
	})
	if err != nil {
		return nil, err
	}
	if mysqlFactory == nil {
		return nil, fmt.Errorf("mysql factory 未初始化")
	}
	return mysqlFactory, nil
}

// autoMigrate 自动建表(与 configs/schema/news_ai_agent.sql 等价)
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.NewsTaskModel{},
		&model.NewsSourceArticleModel{},
		&model.NewsArticleModel{},
		&model.NewsDraftModel{},
	)
}

func (d *datastore) NewsTaskRepository() repository.NewsTaskRepository {
	if d.newsTaskRepository == nil {
		d.newsTaskRepository = &newsTaskRepository{db: d.db}
	}
	return d.newsTaskRepository
}

func (d *datastore) NewsSourceArticleRepository() repository.NewsSourceArticleRepository {
	if d.newsSourceArticleRepository == nil {
		d.newsSourceArticleRepository = &newsSourceArticleRepository{db: d.db}
	}
	return d.newsSourceArticleRepository
}

func (d *datastore) NewsArticleRepository() repository.NewsArticleRepository {
	if d.newsArticleRepository == nil {
		d.newsArticleRepository = &newsArticleRepository{db: d.db}
	}
	return d.newsArticleRepository
}

func (d *datastore) NewsDraftRepository() repository.NewsDraftRepository {
	if d.newsDraftRepository == nil {
		d.newsDraftRepository = &newsDraftRepository{db: d.db}
	}
	return d.newsDraftRepository
}

func (d *datastore) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// dbWithCtx 统一携带 context
func dbWithCtx(ctx context.Context, db *gorm.DB) *gorm.DB { return db.WithContext(ctx) }
