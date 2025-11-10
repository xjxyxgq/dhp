package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// 数据库配置
const (
	DBHost     = "localhost"
	DBPort     = "3306"
	DBUser     = "root"
	DBPassword = "your_password"
	DBName     = "your_database"

	// 表名前缀，实际表名为 t_p_xxx_99_00 ~ t_p_xxx_99_09
	TablePrefix = "t_p_xxx_99_"
)

func main() {
	// 1. 连接数据库
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		DBUser, DBPassword, DBHost, DBPort, DBName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 测试数据库连接
	if err := db.Ping(); err != nil {
		log.Fatalf("数据库连接不可用: %v", err)
	}
	fmt.Println("数据库连接成功")

	// 2. 读取 ID 列表（从文件读取，也可以直接在代码中定义）
	ids := readIDsFromFile("id_list.txt")
	// 或者直接定义 ID 列表：
	// ids := []string{"123", "456", "789"}

	if len(ids) == 0 {
		log.Fatal("ID 列表为空")
	}
	fmt.Printf("共读取 %d 个 ID\n\n", len(ids))

	// 3. 遍历每个 ID，查询所有分表
	for _, id := range ids {
		fmt.Printf("==================== 查询 ID: %s ====================\n", id)
		found := false

		// 遍历 10 个分表（00-09）
		for i := 0; i < 10; i++ {
			tableName := fmt.Sprintf("%s%02d", TablePrefix, i)
			exists, err := checkIDExists(db, tableName, id)

			if err != nil {
				fmt.Printf("  [错误] 表 %s 查询失败: %v\n", tableName, err)
				continue
			}

			if exists {
				fmt.Printf("  [✓] 在表 %s 中找到\n", tableName)
				found = true
			}
		}

		if !found {
			fmt.Printf("  [✗] 在所有表中都未找到\n")
		}
		fmt.Println()
	}
}

// 检查 ID 是否存在于指定表中
func checkIDExists(db *sql.DB, tableName, id string) (bool, error) {
	query := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE id = ?", tableName)

	var count int
	err := db.QueryRow(query, id).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// 从文件读取 ID 列表（每行一个 ID）
func readIDsFromFile(filename string) []string {
	file, err := os.Open(filename)
	if err != nil {
		// 如果文件不存在，返回空列表
		fmt.Printf("警告: 无法打开文件 %s: %v\n", filename, err)
		fmt.Println("提示: 请在代码中直接定义 ID 列表，或创建 id_list.txt 文件\n")
		return []string{}
	}
	defer file.Close()

	var ids []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") { // 忽略空行和注释
			ids = append(ids, line)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("读取文件错误: %v", err)
	}

	return ids
}
