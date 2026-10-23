// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: HTTP API 服务配置

package options

import "fmt"

// ServingOptions REST API 服务配置项
type ServingOptions struct {
	BindAddress string `mapstructure:"bind-address"`
	BindPort    int    `mapstructure:"bind-port"`
}

// NewServingOptions 默认值
func NewServingOptions() *ServingOptions {
	return &ServingOptions{BindAddress: "0.0.0.0", BindPort: 5061}
}

// Validate 参数校验
func (o *ServingOptions) Validate() []error {
	errs := make([]error, 0)
	if o.BindPort <= 0 || o.BindPort > 65535 {
		errs = append(errs, fmt.Errorf("serving.bind-port 非法: %d", o.BindPort))
	}
	return errs
}

// Address 监听地址
func (o *ServingOptions) Address() string {
	return fmt.Sprintf("%s:%d", o.BindAddress, o.BindPort)
}
