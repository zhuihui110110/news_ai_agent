// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 日志配置

package options

import "fmt"

// LogOptions 日志配置项
type LogOptions struct {
	Level    string `mapstructure:"level"`
	Director string `mapstructure:"director"`
}

// NewLogOptions 默认值
func NewLogOptions() *LogOptions {
	return &LogOptions{Level: "debug", Director: "log/runtime"}
}

// Validate 参数校验
func (o *LogOptions) Validate() []error {
	errs := make([]error, 0)
	switch o.Level {
	case "debug", "info", "warn", "error":
	default:
		errs = append(errs, fmt.Errorf("log.level 非法: %s", o.Level))
	}
	return errs
}
