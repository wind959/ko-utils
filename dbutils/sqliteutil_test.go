package dbutils

import (
	"fmt"
	"github.com/wind959/ko-utils/dbutils/sqliteutil"
	"log"
	"testing"
)

func TestSqlite(t *testing.T) {
	err := sqliteutil.InitSqliteDB("test.db")
	if err != nil {
		t.Errorf("InitSqliteDB() error = %v", err)
		return
	}
	defer sqliteutil.Close()

	// 2. 创建表
	columns := map[string]string{
		"id":      "INTEGER PRIMARY KEY AUTOINCREMENT",
		"name":    "TEXT NOT NULL",
		"age":     "INTEGER",
		"created": "DATETIME DEFAULT CURRENT_TIMESTAMP",
	}
	err = sqliteutil.CreateTable("users", columns, false)
	if err != nil {
		log.Fatal(err)
	}
	// 3. 插入数据
	data := map[string]interface{}{
		"name": "Alice",
		"age":  25,
	}
	id, err := sqliteutil.Insert("users", data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Inserted record with ID: %d\n", id)

	// 4. 查询数据
	rows, err := sqliteutil.Query("users", []string{"id", "name", "age"}, "age > ?", 20)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, age int
		var name string
		err = rows.Scan(&id, &name, &age)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("ID: %d, Name: %s, Age: %d\n", id, name, age)
	}

	// 5. 更新数据
	updateData := map[string]interface{}{
		"age": 26,
	}
	affected, err := sqliteutil.Update("users", updateData, "name = ?", "Alice")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Updated %d rows\n", affected)

	// 6. 删除数据
	affected, err = sqliteutil.Delete("users", "name = ?", "Alice")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Deleted %d rows\n", affected)
}
