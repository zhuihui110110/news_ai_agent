// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 任务调度配置(参照 stock_ai taskserver 的 task 配置项)

package options

import "fmt"

// TaskOptions 任务配置项
type TaskOptions struct {
	MaxRetries             int  `mapstructure:"max-retries"`
	RetryIntervalSecond    int  `mapstructure:"retry-interval-second"`
	ScanIntervalSecond     int  `mapstructure:"scan-interval-second"`
	AutoCreate             bool `mapstructure:"auto-create"`
	AutoCreateIntervalHour int  `mapstructure:"auto-create-interval-hour"`
}

// NewTaskOptions 默认值
func NewTaskOptions() *TaskOptions {
	return &TaskOptions{
		MaxRetries:             3,
		RetryIntervalSecond:    30,
		ScanIntervalSecond:     30,
		AutoCreate:             true,
		AutoCreateIntervalHour: 6,
	}
}

// Validate 参数校验
func (o *TaskOptions) Validate() []error {
	errs := make([]error, 0)
	if o.MaxRetries < 0 {
		errs = append(errs, fmt.Errorf("task.max-retries 不能为负"))
	}
	if o.ScanIntervalSecond <= 0 {
		errs = append(errs, fmt.Errorf("task.scan-interval-second 必须大于0"))
	}
	return errs
}
