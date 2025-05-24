package sqliteutil

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"sync"
	"time"
)

var (
	dbInstance *sql.DB
	mu         sync.RWMutex // 保护 dbInstance 的并发访问
)

// InitSqliteDB 打开数据库连接
func InitSqliteDB(dataSourceName string) error {
	mu.Lock()
	defer mu.Unlock()

	if dbInstance != nil {
		return errors.New("database is already open")
	}

	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// 配置连接池
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	dbInstance = db
	return nil
}

// Close 关闭数据库连接
func Close() error {
	mu.Lock()
	defer mu.Unlock()

	if dbInstance == nil {
		return errors.New("database is not open")
	}

	err := dbInstance.Close()
	dbInstance = nil
	return err
}

// CreateTable 创建表(优化版)
// tableName: 表名
// columns: 列定义映射(列名:类型)
// overwrite: true=表存在时删除重建，false=表存在时返回错误
func CreateTable(tableName string, columns map[string]string, overwrite bool) error {
	mu.Lock()
	defer mu.Unlock()

	if dbInstance == nil {
		return errors.New("database is not open")
	}

	if len(columns) == 0 {
		return errors.New("no columns provided")
	}

	// 检查表是否存在
	tableExists, err := checkTableExists(tableName)
	if err != nil {
		return fmt.Errorf("failed to check table existence: %v", err)
	}

	if tableExists {
		if !overwrite {
			return fmt.Errorf("table %s already exists", tableName)
		}
		// 只有表存在且overwrite=true时才执行删除
		_, err = dbInstance.Exec(fmt.Sprintf("DROP TABLE %s", tableName))
		if err != nil {
			return fmt.Errorf("failed to drop table: %v", err)
		}
	}

	// 构建创建表SQL
	query := fmt.Sprintf("CREATE TABLE %s (", tableName)
	first := true
	for name, typ := range columns {
		if !first {
			query += ", "
		}
		query += fmt.Sprintf("%s %s", name, typ)
		first = false
	}
	query += ")"

	_, err = dbInstance.Exec(query)
	return err
}

// Insert 插入数据
func Insert(tableName string, data map[string]interface{}) (int64, error) {
	mu.RLock()
	defer mu.RUnlock()

	if dbInstance == nil {
		return 0, errors.New("database is not open")
	}

	if len(data) == 0 {
		return 0, errors.New("no data provided")
	}

	columns := ""
	placeholders := ""
	values := make([]interface{}, 0, len(data))
	i := 1
	for col, val := range data {
		if i > 1 {
			columns += ", "
			placeholders += ", "
		}
		columns += col
		placeholders += fmt.Sprintf("$%d", i)
		values = append(values, val)
		i++
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, columns, placeholders)
	result, err := dbInstance.Exec(query, values...)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// Update 更新数据
func Update(tableName string, data map[string]interface{}, where string, args ...interface{}) (int64, error) {
	mu.RLock()
	defer mu.RUnlock()

	if dbInstance == nil {
		return 0, errors.New("database is not open")
	}

	if len(data) == 0 {
		return 0, errors.New("no data provided")
	}

	setClause := ""
	values := make([]interface{}, 0, len(data)+len(args))
	i := 1
	for col, val := range data {
		if i > 1 {
			setClause += ", "
		}
		setClause += fmt.Sprintf("%s = $%d", col, i)
		values = append(values, val)
		i++
	}

	// 添加 WHERE 条件参数
	for _, arg := range args {
		values = append(values, arg)
	}

	query := fmt.Sprintf("UPDATE %s SET %s", tableName, setClause)
	if where != "" {
		query += " WHERE " + where
	}

	result, err := dbInstance.Exec(query, values...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Delete 删除数据
func Delete(tableName string, where string, args ...interface{}) (int64, error) {
	mu.RLock()
	defer mu.RUnlock()

	if dbInstance == nil {
		return 0, errors.New("database is not open")
	}

	query := fmt.Sprintf("DELETE FROM %s", tableName)
	if where != "" {
		query += " WHERE " + where
	}

	result, err := dbInstance.Exec(query, args...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Query 查询数据
func Query(tableName string, columns []string, where string, args ...interface{}) (*sql.Rows, error) {
	mu.RLock()
	defer mu.RUnlock()

	if dbInstance == nil {
		return nil, errors.New("database is not open")
	}

	cols := "*"
	if len(columns) > 0 {
		cols = ""
		for i, col := range columns {
			if i > 0 {
				cols += ", "
			}
			cols += col
		}
	}

	query := fmt.Sprintf("SELECT %s FROM %s", cols, tableName)
	if where != "" {
		query += " WHERE " + where
	}

	return dbInstance.Query(query, args...)
}

// QueryRow 查询单行数据
func QueryRow(tableName string, columns []string, where string, args ...interface{}) *sql.Row {
	mu.RLock()
	defer mu.RUnlock()

	if dbInstance == nil {
		return nil
	}

	cols := "*"
	if len(columns) > 0 {
		cols = ""
		for i, col := range columns {
			if i > 0 {
				cols += ", "
			}
			cols += col
		}
	}

	query := fmt.Sprintf("SELECT %s FROM %s", cols, tableName)
	if where != "" {
		query += " WHERE " + where
	}

	return dbInstance.QueryRow(query, args...)
}

// Execute 执行原始 SQL
func Execute(query string, args ...interface{}) (sql.Result, error) {
	mu.RLock()
	defer mu.RUnlock()

	if dbInstance == nil {
		return nil, errors.New("database is not open")
	}

	return dbInstance.Exec(query, args...)
}

// BeginTransaction 开始事务
func BeginTransaction() (*sql.Tx, error) {
	mu.RLock()
	defer mu.RUnlock()

	if dbInstance == nil {
		return nil, errors.New("database is not open")
	}

	return dbInstance.Begin()
}

// Ping 检查数据库连接
func Ping() error {
	mu.RLock()
	defer mu.RUnlock()

	if dbInstance == nil {
		return errors.New("database is not open")
	}

	return dbInstance.Ping()
}

// checkTableExists 检查表是否存在(保持不变)
func checkTableExists(tableName string) (bool, error) {
	query := `
		SELECT count(*) 
		FROM sqlite_master 
		WHERE type='table' AND name=?
	`
	var count int
	err := dbInstance.QueryRow(query, tableName).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
