package excelutil

import (
	"fmt"
	"log"
	"testing"
	"time"
)

type User struct {
	ID        int
	Name      string
	Email     string
	Age       int
	Salary    float64
	CreatedAt time.Time
}

func TestExcelUtils(t *testing.T) {
	// 1. 创建新Excel
	excel := NewExcel()
	defer excel.Close()

	// 2. 写入数据
	data := [][]interface{}{
		{1, "张三", "zhangsan@example.com", 25, 8000.50, time.Now()},
		{2, "李四", "lisi@example.com", 30, 12000.00, time.Now()},
	}

	headers := []string{"ID", "姓名", "邮箱", "年龄", "薪资", "创建时间"}

	opt := WriteOption{
		Sheet:    "用户数据",
		Headers:  headers,
		StartRow: 1,
	}

	if err := excel.Write(data, opt); err != nil {
		log.Fatal(err)
	}

	// 3. 保存文件
	if err := excel.SaveAs("users.xlsx"); err != nil {
		log.Fatal(err)
	}

	// 4. 读取数据
	excel2, err := Open("users.xlsx")
	if err != nil {
		log.Fatal(err)
	}
	defer excel2.Close()

	readOpt := ReadOption{
		Sheet:     "用户数据",
		HasHeader: true,
		StartRow:  2, // 跳过表头
	}

	var users []User
	if err := excel2.ReadToSlice(&users, readOpt); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("读取到 %d 条数据\n", len(users))
	for _, user := range users {
		fmt.Printf("ID: %d, Name: %s, Email: %s\n", user.ID, user.Name, user.Email)
	}

	// 5. 使用便捷函数
	users2 := []User{
		{1, "王五", "wangwu@example.com", 28, 9000.00, time.Now()},
		{2, "赵六", "zhaoliu@example.com", 35, 15000.00, time.Now()},
	}

	// 一键导出
	if err := ExportStructs("users2.xlsx", users2, headers); err != nil {
		log.Fatal(err)
	}

	// 一键导入
	var importedUsers []User
	if err := ImportFromFile("users2.xlsx", &importedUsers); err != nil {
		log.Fatal(err)
	}
}
