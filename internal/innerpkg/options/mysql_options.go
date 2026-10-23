// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: MySQL 连接配置(参照 stock_ai genericoptions.MySQLOptions 精简)

package options

import (
	"fmt"
	"time"
)

// MySQLOptions MySQL 配置项
type MySQLOptions struct {
	Host        string        `mapstructure:"host"`
	Port        int           `mapstructure:"port"`
	Username    string        `mapstructure:"username"`
	Password    string        `mapstructure:"password"`
	Database    string        `mapstructure:"database"`
	Charset     string        `mapstructure:"charset"`
	MaxIdleConn int           `mapstructure:"max-idle-conn"`
	MaxOpenConn int           `mapstructure:"max-open-conn"`
	LogLevel    int           `mapstructure:"log-level"`
	Migrate     bool          `mapstructure:"migrate"`
	Timeout     time.Duration `mapstructure:"timeout"`
}

// NewMySQLOptions 默认值
func NewMySQLOptions() *MySQLOptions {
	return &MySQLOptions{
		Host:        "127.0.0.1",
		Port:        3306,
		Username:    "root",
		Database:    "news_ai_agent",
		Charset:     "utf8mb4",
		MaxIdleConn: 10,
		MaxOpenConn: 50,
		LogLevel:    4,
		Migrate:     true,
		Timeout:     10 * time.Second,
	}
}

// Validate 参数校验
func (o *MySQLOptions) Validate() []error {
	errs := make([]error, 0)
	if o.Host == "" {
		errs = append(errs, fmt.Errorf("mysql.host 不能为空"))
	}
	if o.Port <= 0 {
		errs = append(errs, fmt.Errorf("mysql.port 必须大于0"))
	}
	if o.Database == "" {
		errs = append(errs, fmt.Errorf("mysql.database 不能为空"))
	}
	return errs
}

// DSN 生成 gorm DSN
func (o *MySQLOptions) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=Local&timeout=%s",
		o.Username, o.Password, o.Host, o.Port, o.Database, o.Charset, o.Timeout)
}
