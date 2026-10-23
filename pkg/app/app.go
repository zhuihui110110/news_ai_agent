// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 应用启动框架(参照 stock_ai pkg/app, 基于 cobra+viper 精简实现:
// 统一 --config 配置加载, 环境变量覆盖, Complete/Validate/Run 生命周期)

package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Option 应用可选参数(函数式选项)
type Option func(*App)

// WithOptions 注入业务 Options 对象(需实现 FlagSet 接口)
func WithOptions(o Options) Option {
	return func(a *App) { a.options = o }
}

// WithRunFunc 注入真正的运行逻辑
func WithRunFunc(run RunFunc) Option {
	return func(a *App) { a.runFunc = run }
}

// WithDescription 应用描述
func WithDescription(desc string) Option {
	return func(a *App) { a.description = desc }
}

// Options 业务配置聚合接口(由各服务 soptions.Options 实现, 配置来自 yaml+环境变量)
type Options interface {
	Validate() []error
	Complete() error
	String() string
}

// RunFunc 运行函数, 入参为服务基础名
type RunFunc func(basename string) error

// App 命令行应用
type App struct {
	basename    string
	description string
	options     Options
	runFunc     RunFunc
	rootCmd     *cobra.Command
}

// NewApp 创建应用
func NewApp(name string, opts ...Option) *App {
	a := &App{basename: name}
	for _, o := range opts {
		o(a)
	}
	cmd := cobra.Command{
		Use:     FormatBaseName(name),
		Short:   a.description,
		SilenceUsage: true,
		RunE:    a.runCommand,
	}
	cmd.PersistentFlags().StringP("config", "c", fmt.Sprintf("configs/%s.yml", name), "配置文件路径")
	a.rootCmd = &cmd
	return a
}

// Run 启动应用
func (a *App) Run() {
	if err := a.rootCmd.Execute(); err != nil {
		fmt.Printf("app run failed: %v\n", err)
		os.Exit(1)
	}
}

func (a *App) runCommand(cmd *cobra.Command, args []string) error {
	printWorkingDir()
	printFlags(cmd, a.basename)

	// 配置文件 + 环境变量(前缀 = 服务名大写, . 或 - 转 _)覆盖
	cfg, _ := cmd.Flags().GetString("config")
	if err := loadConfig(cfg, a.basename); err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}
	if a.options != nil {
		if err := viper.Unmarshal(a.options); err != nil {
			return fmt.Errorf("解析配置到 Options 失败: %w", err)
		}
		if err := a.options.Complete(); err != nil {
			return fmt.Errorf("Complete 失败: %w", err)
		}
		if errs := a.options.Validate(); len(errs) > 0 {
			for _, e := range errs {
				fmt.Printf("参数校验失败: %v\n", e)
			}
			return fmt.Errorf("参数校验失败")
		}
	}
	if a.runFunc != nil {
		return a.runFunc(a.basename)
	}
	return nil
}

func loadConfig(path, basename string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	viper.SetConfigFile(abs)
	viper.AutomaticEnv()
	viper.SetEnvPrefix(strings.ToUpper(basename))
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件 %s 失败: %w", abs, err)
	}
	fmt.Printf("配置文件: %s\n", abs)
	return nil
}

func printWorkingDir() {
	wd, _ := os.Getwd()
	fmt.Printf("working directory: %s\n", wd)
}

func printFlags(cmd *cobra.Command, name string) {
	fmt.Printf("%s flags:\n", name)
	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		fmt.Printf("  --%s: %s\n", f.Name, f.DefValue)
	})
}

// FormatBaseName 服务二进制基础名(去掉路径)
func FormatBaseName(basename string) string {
	basename = filepath.Base(basename)
	return strings.ReplaceAll(basename, "_", "-")
}
