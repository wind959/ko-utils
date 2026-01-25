package dbutils

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/wind959/ko-utils/dbutils/sqliteutil"
)

func TestSqlite(t *testing.T) {
	// 初始化
	db := sqliteutil.GetInstance()
	err := db.Init("test.db",
		sqliteutil.WithPoolSize(10, 5),
		sqliteutil.WithConnLifetime(10*time.Minute, 5*time.Minute),
	)

	if err != nil {
		fmt.Println(err)
	}

	// 创建表
	columns := map[string]string{
		"id":         "INTEGER",
		"name":       "TEXT",
		"age":        "INTEGER",
		"created_at": "DATETIME",
	}
	err = db.CreateTable(context.Background(), "users", columns,
		sqliteutil.WithPrimaryKey("id"),
		sqliteutil.WithUniqueConstraint("name"),
	)

	// 插入数据
	_, err = db.Insert(context.Background(), "users", map[string]interface{}{
		"name":       "Alice",
		"age":        25,
		"created_at": time.Now(),
	})

	// 查询数据
	rows, err := db.Query(context.Background(), "users", []string{"id", "name"}, "age > ?", 20)
	defer rows.Close()

	// 使用事务
	err = db.Transaction(context.Background(), func(tx *sql.Tx) error {
		// 在事务中执行多个操作
		_, err := tx.Exec("UPDATE users SET age = age + 1 WHERE id = ?", 1)
		return err
	})
}
