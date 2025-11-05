package viperutil

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"strings"
	"time"
)

// Config Viper 配置管理器
type Config struct {
	viper *viper.Viper
}

// NewConfig 创建 Config 实例
func NewConfig() *Config {
	v := viper.New()
	v.AutomaticEnv()                                   // 自动加载环境变量
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // 将环境变量中的 . 替换为 _
	return &Config{viper: v}
}

// SetConfigFile 设置配置文件路径
func (c *Config) SetConfigFile(path string) {
	c.viper.SetConfigFile(path)
}

// SetConfigType 设置配置文件类型（如 "json", "yaml" 等）
func (c *Config) SetConfigType(configType string) {
	c.viper.SetConfigType(configType)
}

// ReadConfig 读取配置文件
func (c *Config) ReadConfig() error {
	if err := c.viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}
	return nil
}

// Get 获取任意类型配置值
func (c *Config) Get(key string) interface{} {
	return c.viper.Get(key)
}

// GetString 获取字符串配置值
func (c *Config) GetString(key string) string {
	return c.viper.GetString(key)
}

// GetInt 获取整数配置值
func (c *Config) GetInt(key string) int {
	return c.viper.GetInt(key)
}

// GetInt32 获取 32 位整数配置值
func (c *Config) GetInt32(key string) int32 {
	return c.viper.GetInt32(key)
}

// GetInt64 获取 64 位整数配置值
func (c *Config) GetInt64(key string) int64 {
	return c.viper.GetInt64(key)
}

// GetBool 获取布尔配置值
func (c *Config) GetBool(key string) bool {
	return c.viper.GetBool(key)
}

// GetFloat64 获取浮点数配置值
func (c *Config) GetFloat64(key string) float64 {
	return c.viper.GetFloat64(key)
}

// GetStringSlice 获取字符串切片配置值
func (c *Config) GetStringSlice(key string) []string {
	return c.viper.GetStringSlice(key)
}

// GetUintSlice 获取无符号整型切片类型配置值
func (c *Config) GetUintSlice(key string) []uint {
	// 首先尝试直接获取uint切片
	if value := c.viper.Get(key); value != nil {
		if uintSlice, ok := value.([]uint); ok {
			return uintSlice
		}
	}
	// 如果直接获取失败，尝试从int切片转换
	intSlice := c.viper.GetIntSlice(key)
	if len(intSlice) == 0 {
		return []uint{}
	}
	// 将int切片转换为uint切片
	uintSlice := make([]uint, len(intSlice))
	for i, v := range intSlice {
		if v < 0 {
			// 如果值为负数，设置为0
			uintSlice[i] = 0
		} else {
			uintSlice[i] = uint(v)
		}
	}
	return uintSlice
}

// GetStringMap 获取字符串映射配置值
func (c *Config) GetStringMap(key string) map[string]interface{} {
	return c.viper.GetStringMap(key)
}

// GetDuration 获取时间段类型配置值
func (c *Config) GetDuration(key string) time.Duration {
	return c.viper.GetDuration(key)
}

// SetDefault 设置默认配置值
func (c *Config) SetDefault(key string, value interface{}) {
	c.viper.SetDefault(key, value)
}

// BindEnv 绑定环境变量
func (c *Config) BindEnv(key string) error {
	return c.viper.BindEnv(key)
}

// Unmarshal 将配置反序列化为结构体
func (c *Config) Unmarshal(v interface{}) error {
	return c.viper.Unmarshal(v)
}

// WatchConfig 监听配置文件变化
func (c *Config) WatchConfig() {
	c.viper.WatchConfig()
}

// OnConfigChange 注册配置文件变化回调函数
func (c *Config) OnConfigChange(fn func(in fsnotify.Event)) {
	c.viper.OnConfigChange(fn)
}

// Set 设置配置值
func (c *Config) Set(key string, value interface{}) {
	c.viper.Set(key, value)
}

// SaveConfig 保存配置到文件
func (c *Config) SaveConfig() error {
	// 保存当前的配置值
	currentSettings := c.viper.AllSettings()

	// 重新读取配置文件，确保所有原始配置项都被加载
	if err := c.viper.ReadInConfig(); err != nil {
		return err
	}
	// 将修改后的配置项重新设置到 viper 中
	for key, value := range currentSettings {
		c.viper.Set(key, value)
	}
	// 写入配置
	err := c.viper.WriteConfig()
	if err != nil {
		return err
	}
	return nil
}
