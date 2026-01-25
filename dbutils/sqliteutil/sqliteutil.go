package sqliteutil

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	instance *DBManager
	once     sync.Once
)

type DBManager struct {
	db *sql.DB
	mu sync.RWMutex // 读写锁：读操作可以并发，写操作独占
}

// Config 数据库配置
type Config struct {
	MaxOpenConns    int           // 最大打开连接数
	MaxIdleConns    int           // 最大空闲连接数
	ConnMaxLifetime time.Duration // 连接最大生命周期
	ConnMaxIdleTime time.Duration // 连接最大空闲时间
}

// Option 配置选项函数
type Option func(*Config)

// GetInstance 获取数据库管理器单例
func GetInstance() *DBManager {
	once.Do(func() {
		instance = &DBManager{}
	})
	return instance
}

// TableOptions 表选项
type TableOptions struct {
	Overwrite     bool
	PrimaryKey    string
	UniqueColumns [][]string
}

// TableOption 表选项函数
type TableOption func(*TableOptions)

// Init 初始化数据库连接
func (m *DBManager) Init(dataSourceName string, options ...Option) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.db != nil {
		return errors.New("database already initialized")
	}

	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// 应用配置选项
	config := &Config{
		MaxOpenConns:    25,
		MaxIdleConns:    25,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}
	for _, opt := range options {
		opt(config)
	}

	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)
	m.db = db

	// 测试连接
	if err := m.db.Ping(); err != nil {
		m.db.Close()
		m.db = nil
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// Close 关闭数据库连接
func (m *DBManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.db == nil {
		return errors.New("database not initialized")
	}

	err := m.db.Close()
	m.db = nil
	return err
}

// Ping 检查数据库连接
func (m *DBManager) Ping(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.db == nil {
		return errors.New("database not initialized")
	}

	return m.db.PingContext(ctx)
}

// CreateTable 创建表
func (m *DBManager) CreateTable(ctx context.Context, tableName string, columns map[string]string, options ...TableOption) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.db == nil {
		return errors.New("database not initialized")
	}

	if len(columns) == 0 {
		return errors.New("no columns provided")
	}

	// 应用表选项
	tableOpts := &TableOptions{
		Overwrite:  false,
		PrimaryKey: "",
	}
	for _, opt := range options {
		opt(tableOpts)
	}
	// 检查表是否存在
	exists, err := m.tableExists(ctx, tableName)
	if err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if exists {
		if !tableOpts.Overwrite {
			return fmt.Errorf("table %s already exists", tableName)
		}
		// 删除现有表
		_, err = m.db.ExecContext(ctx, fmt.Sprintf("DROP TABLE %s", tableName))
		if err != nil {
			return fmt.Errorf("failed to drop table: %w", err)
		}
	}

	// 构建列定义
	var columnDefs []string
	for name, typ := range columns {
		columnDefs = append(columnDefs, fmt.Sprintf("%s %s", name, typ))
	}

	// 添加主键约束
	if tableOpts.PrimaryKey != "" {
		columnDefs = append(columnDefs, fmt.Sprintf("PRIMARY KEY (%s)", tableOpts.PrimaryKey))
	}

	// 添加唯一约束
	for _, uniqueCol := range tableOpts.UniqueColumns {
		columnDefs = append(columnDefs, fmt.Sprintf("UNIQUE (%s)", strings.Join(uniqueCol, ", ")))
	}

	// 创建表
	query := fmt.Sprintf("CREATE TABLE %s (\n  %s\n)", tableName, strings.Join(columnDefs, ",\n  "))
	_, err = m.db.ExecContext(ctx, query)
	return err
}

// TableExists 检查表是否存在
func (m *DBManager) TableExists(ctx context.Context, tableName string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.tableExists(ctx, tableName)
}

// tableExists 检查表是否存在
func (m *DBManager) tableExists(ctx context.Context, tableName string) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM sqlite_master 
		WHERE type='table' AND name=?
	`
	var count int
	err := m.db.QueryRowContext(ctx, query, tableName).Scan(&count)
	return count > 0, err
}

// Insert 插入单条数据
func (m *DBManager) Insert(ctx context.Context, tableName string, data map[string]interface{}) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.db == nil {
		return 0, errors.New("database not initialized")
	}

	if len(data) == 0 {
		return 0, errors.New("no data provided")
	}

	columns := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))

	i := 1
	for col, val := range data {
		columns = append(columns, col)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		values = append(values, val)
		i++
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	result, err := m.db.ExecContext(ctx, query, values...)
	if err != nil {
		return 0, fmt.Errorf("insert failed: %w", err)
	}

	return result.LastInsertId()
}

// BatchInsert 批量插入数据
func (m *DBManager) BatchInsert(ctx context.Context, tableName string, columns []string, data [][]interface{}) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.db == nil {
		return 0, errors.New("database not initialized")
	}

	if len(data) == 0 {
		return 0, errors.New("no data provided")
	}

	// 开始事务
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin transaction failed: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 准备插入语句
	placeholders := make([]string, len(columns))
	for i := range columns {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("prepare statement failed: %w", err)
	}
	defer stmt.Close()

	var totalRows int64
	for _, row := range data {
		result, err := stmt.ExecContext(ctx, row...)
		if err != nil {
			return totalRows, fmt.Errorf("batch insert failed: %w", err)
		}
		rows, _ := result.RowsAffected()
		totalRows += rows
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return totalRows, fmt.Errorf("commit failed: %w", err)
	}

	return totalRows, nil
}

// Update 更新数据
func (m *DBManager) Update(ctx context.Context, tableName string, data map[string]interface{}, where string, args ...interface{}) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.db == nil {
		return 0, errors.New("database not initialized")
	}

	if len(data) == 0 {
		return 0, errors.New("no data provided")
	}

	setClause := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))

	i := 1
	for col, val := range data {
		setClause = append(setClause, fmt.Sprintf("%s = $%d", col, i))
		values = append(values, val)
		i++
	}

	// 添加 WHERE 条件参数
	values = append(values, args...)

	query := fmt.Sprintf("UPDATE %s SET %s", tableName, strings.Join(setClause, ", "))
	if where != "" {
		query += " WHERE " + where
	}

	result, err := m.db.ExecContext(ctx, query, values...)
	if err != nil {
		return 0, fmt.Errorf("update failed: %w", err)
	}

	return result.RowsAffected()
}

// Delete 删除数据
func (m *DBManager) Delete(ctx context.Context, tableName string, where string, args ...interface{}) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.db == nil {
		return 0, errors.New("database not initialized")
	}

	query := fmt.Sprintf("DELETE FROM %s", tableName)
	if where != "" {
		query += " WHERE " + where
	}

	result, err := m.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("delete failed: %w", err)
	}

	return result.RowsAffected()
}

// Query 查询数据
func (m *DBManager) Query(ctx context.Context, tableName string, columns []string, where string, args ...interface{}) (*sql.Rows, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.db == nil {
		return nil, errors.New("database not initialized")
	}

	cols := "*"
	if len(columns) > 0 {
		cols = strings.Join(columns, ", ")
	}

	query := fmt.Sprintf("SELECT %s FROM %s", cols, tableName)
	if where != "" {
		query += " WHERE " + where
	}

	return m.db.QueryContext(ctx, query, args...)
}

// QueryRow 查询单行数据
func (m *DBManager) QueryRow(ctx context.Context, tableName string, columns []string, where string, args ...interface{}) *sql.Row {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.db == nil {
		return nil
	}

	cols := "*"
	if len(columns) > 0 {
		cols = strings.Join(columns, ", ")
	}

	query := fmt.Sprintf("SELECT %s FROM %s", cols, tableName)
	if where != "" {
		query += " WHERE " + where
	}

	return m.db.QueryRowContext(ctx, query, args...)
}

// QueryWithCallback 查询数据并使用回调处理每一行
func (m *DBManager) QueryWithCallback(ctx context.Context, tableName string, columns []string, where string,
	callback func(*sql.Rows) error, args ...interface{}) error {

	rows, err := m.Query(ctx, tableName, columns, where, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		if err := callback(rows); err != nil {
			return err
		}
	}

	return rows.Err()
}

// BeginTx 开始事务
func (m *DBManager) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.db == nil {
		return nil, errors.New("database not initialized")
	}

	return m.db.BeginTx(ctx, opts)
}

// Transaction 执行事务操作
func (m *DBManager) Transaction(ctx context.Context, fn func(*sql.Tx) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.db == nil {
		return errors.New("database not initialized")
	}

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // 重新抛出panic
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}

// Execute 执行原始 SQL
func (m *DBManager) Execute(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.db == nil {
		return nil, errors.New("database not initialized")
	}

	return m.db.ExecContext(ctx, query, args...)
}

// GetTableInfo 获取表结构信息
func (m *DBManager) GetTableInfo(ctx context.Context, tableName string) ([]ColumnInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.db == nil {
		return nil, errors.New("database not initialized")
	}

	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		err := rows.Scan(
			&col.CID,
			&col.Name,
			&col.Type,
			&col.NotNull,
			&col.DefaultValue,
			&col.PK,
		)
		if err != nil {
			return nil, err
		}
		columns = append(columns, col)
	}

	return columns, nil
}

// Vacuum 清理数据库碎片
func (m *DBManager) Vacuum(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.db == nil {
		return errors.New("database not initialized")
	}

	_, err := m.db.ExecContext(ctx, "VACUUM")
	return err
}

// WithPoolSize 设置连接池大小
func WithPoolSize(maxOpen, maxIdle int) Option {
	return func(c *Config) {
		c.MaxOpenConns = maxOpen
		c.MaxIdleConns = maxIdle
	}
}

// WithConnLifetime 设置连接生命周期
func WithConnLifetime(lifetime, idleTime time.Duration) Option {
	return func(c *Config) {
		c.ConnMaxLifetime = lifetime
		c.ConnMaxIdleTime = idleTime
	}
}

// WithOverwrite 设置是否覆盖已存在的表
func WithOverwrite(overwrite bool) TableOption {
	return func(o *TableOptions) {
		o.Overwrite = overwrite
	}
}

// WithPrimaryKey 设置主键
func WithPrimaryKey(key string) TableOption {
	return func(o *TableOptions) {
		o.PrimaryKey = key
	}
}

// WithUniqueConstraint 添加唯一约束
func WithUniqueConstraint(columns ...string) TableOption {
	return func(o *TableOptions) {
		o.UniqueColumns = append(o.UniqueColumns, columns)
	}
}

// ColumnInfo 列信息
type ColumnInfo struct {
	CID          int
	Name         string
	Type         string
	NotNull      int
	DefaultValue interface{}
	PK           int
}

// PrintTable 打印表内容（用于调试）
func (m *DBManager) PrintTable(ctx context.Context, tableName string, limit int) error {
	rows, err := m.Query(ctx, tableName, nil, "", nil)
	if err != nil {
		return err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	fmt.Printf("\n=== Table: %s ===\n", tableName)
	fmt.Println(strings.Join(columns, "\t"))

	count := 0
	for rows.Next() {
		if limit > 0 && count >= limit {
			fmt.Printf("... (showing %d rows)\n", limit)
			break
		}

		err := rows.Scan(valuePtrs...)
		if err != nil {
			return err
		}

		for i, val := range values {
			switch v := val.(type) {
			case []byte:
				fmt.Print(string(v))
			case nil:
				fmt.Print("NULL")
			default:
				fmt.Print(v)
			}
			if i < len(values)-1 {
				fmt.Print("\t")
			}
		}
		fmt.Println()
		count++
	}

	fmt.Printf("Total rows: %d\n", count)
	return nil
}

// SafeExec 安全执行SQL，错误时记录日志但不panic
func (m *DBManager) SafeExec(ctx context.Context, query string, args ...interface{}) {
	_, err := m.Execute(ctx, query, args...)
	if err != nil {
		log.Printf("SafeExec failed: %v, query: %s", err, query)
	}
}
