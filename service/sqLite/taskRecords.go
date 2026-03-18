package sqLite

import (
	"fmt"
	"planA/initialization/golabl"
	"planA/tool"
	sqLiteType "planA/type/sqLite"
	"strings"
	"time"
)

// CreateTaskIdTab 创建task_records表
// @return error 错误信息
func CreateTaskIdTab() error {
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS task_records (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id VARCHAR(100) NOT NULL,
        shop_id VARCHAR(100) NOT NULL,
        task_id VARCHAR(100) NOT NULL,
        shop_name VARCHAR(100) NOT NULL,
        is_export INTEGER NOT NULL DEFAULT 0,
        task_type INTEGER NOT NULL DEFAULT 0,
        create_at DATETIME NOT NULL
    );
    
    CREATE INDEX IF NOT EXISTS idx_task_id ON task_records(task_id);
    CREATE INDEX IF NOT EXISTS idx_create_at ON task_records(create_at);
    `
	_, err := golabl.SqliteDb.Exec(createTableSQL)
	if err != nil {
		return err
	}
	return nil
}

// CreateTaskRecords 向task_records表插入一条记录
// @param record TaskRecord 要插入的记录
// @return error 错误信息
func CreateTaskRecords(record sqLiteType.TaskRecords) error {
	// 在 Go 代码中计算当前时间
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	insertSQL := `INSERT INTO task_records (user_id,shop_id,task_id,shop_name,task_type, create_at) VALUES (?, ?, ?, ?, ?, ?)`
	result, err := golabl.SqliteDb.Exec(insertSQL, record.UserID, record.ShopID, record.TaskID, record.ShopName, record.TaskType, currentTime)
	if err != nil {
		return err
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	record.ID = lastID
	return nil
}

// GetTaskRecordsList 分页查询task_records表记录
// @param page 页码（从1开始）
// @param pageSize 每页条数
// @param taskId 任务ID（可选，为空时不作为条件）
// @param shopName 店铺名称（可选，为空时不作为条件）
// @param taskType 任务类型（可选，为空时不作为条件）
// @return []TaskRecord 记录列表
// @return int64 总记录数
// @return error 错误信息
func GetTaskRecordsList(params sqLiteType.GetTaskRecordsByUserIdParams) ([]sqLiteType.TaskRecords, int64, error) {
	// 参数校验
	pageSize, offset := tool.GetPage(params.Page.PageNum, params.Page.PageSize)

	// 构建查询条件
	var conditions []string
	var args []interface{}

	if params.TaskID != "" {
		conditions = append(conditions, "task_id = ?")
		args = append(args, params.TaskID)
	}

	if params.ShopName != "" {
		conditions = append(conditions, "shop_name = ?")
		args = append(args, params.ShopName)
	}
	if params.TaskType != 0 {
		conditions = append(conditions, "task_type = ?")
		args = append(args, params.TaskType)
	}

	// 构建 WHERE子句
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 查询总数
	var total int64
	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM task_records %s`, whereClause)
	var countErr error
	if len(args) > 0 {
		countErr = golabl.SqliteDb.QueryRow(countSQL, args...).Scan(&total)
	} else {
		countErr = golabl.SqliteDb.QueryRow(countSQL).Scan(&total)
	}

	if countErr != nil {
		return nil, 0, fmt.Errorf("查询总数失败: %v", countErr)
	}

	// 分页查询
	querySQL := fmt.Sprintf(`
        SELECT id,user_id, shop_id, task_id, shop_name,is_export, task_type,create_at 
        FROM task_records 
        %s
        ORDER BY id DESC 
        LIMIT ? OFFSET ?
    `, whereClause)

	// 添加分页参数到 args
	queryArgs := append(args, pageSize, offset)
	rows, err := golabl.SqliteDb.Query(querySQL, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询失败: %v", err)
	}
	defer rows.Close()

	var records []sqLiteType.TaskRecords

	for rows.Next() {
		var record sqLiteType.TaskRecords
		err = rows.Scan(&record.ID, &record.UserID, &record.ShopID, &record.TaskID, &record.ShopName, &record.IsExport, &record.TaskType, &record.CreateAt)
		if err != nil {
			return nil, 0, fmt.Errorf("扫描数据失败: %v", err)
		}
		records = append(records, record)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("遍历结果集错误: %v", err)
	}

	return records, total, nil
}

// UpdateTaskRecord 根据任务ID更新任务记录
func UpdateTaskRecord(record sqLiteType.TaskRecords) error {
	updateSQL := `UPDATE task_records SET user_id = ?, shop_id = ?, task_id = ?, shop_name = ?, task_type = ?, is_export = ? WHERE id = ?`
	_, err := golabl.SqliteDb.Exec(updateSQL, record.UserID, record.ShopID, record.TaskID, record.ShopName, record.TaskType, record.IsExport, record.ID)
	return err
}

// DeleteTaskRecordsByTaskID 根据任务ID删除数据
// @param taskID 任务ID
// @return error 错误
func DeleteTaskRecordsByTaskID(taskID string) error {
	_, err := golabl.SqliteDb.Exec("DELETE FROM task_records WHERE task_id = ?", taskID)
	return err
}

// GetTaskRecordByTaskID 根据taskId查询单个任务记录
// @param taskID 任务ID
// @return *TaskRecord 记录指针
// @return error 错误信息
func GetTaskRecordByTaskID(taskID string) (*sqLiteType.TaskRecords, error) {
	query := `SELECT id,user_id, shop_id, task_id, shop_name, task_type, create_at 
              FROM task_records 
              WHERE task_id = ? 
              LIMIT 1`

	var record sqLiteType.TaskRecords
	var createAtStr string

	err := golabl.SqliteDb.QueryRow(query, taskID).Scan(
		&record.ID,
		&record.UserID,
		&record.ShopID,
		&record.TaskID,
		&record.ShopName,
		&record.TaskType,
		&createAtStr,
	)

	if err != nil {
		return nil, fmt.Errorf("查询失败: %v", err)
	}

	return &record, nil
}

// DeleteOldTaskRecords 删除task_records表中3天前的记录
// @return error 错误信息
func DeleteOldTaskRecords() error {
	// 使用SQLite的date函数计算3天前
	result, err := golabl.SqliteDb.Exec(`
        DELETE FROM task_records 
        WHERE create_at < datetime('now', '-3 days')
    `)
	if err != nil {
		return fmt.Errorf("删除旧数据失败: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %v", err)
	}

	fmt.Printf("已删除 %d 条大于3天的记录\n", rowsAffected)
	return nil
}

// GetTaskRecords24Hour 查询task_records中24小时内的所有数据
func GetTaskRecords24Hour() ([]sqLiteType.TaskRecords, error) {
	// 查询24小时内的记录，按创建时间倒序排列
	querySQL := `
    SELECT id, user_id, shop_id, task_id, shop_name, is_export, task_type, create_at 
    FROM task_records 
    WHERE create_at >= datetime('now', '-4 hours')
    AND create_at <= datetime('now', '-10 minutes')
    ORDER BY create_at DESC`

	// 执行查询
	rows, err := golabl.SqliteDb.Query(querySQL)
	if err != nil {
		return nil, fmt.Errorf("查询24小时内任务记录失败: %v", err)
	}
	defer rows.Close() // 确保结果集最终被关闭

	// 初始化结果切片
	var records []sqLiteType.TaskRecords

	// 遍历查询结果
	for rows.Next() {
		var record sqLiteType.TaskRecords
		// 扫描每一行数据到结构体中
		err = rows.Scan(
			&record.ID,
			&record.UserID,
			&record.ShopID,
			&record.TaskID,
			&record.ShopName,
			&record.IsExport,
			&record.TaskType,
			&record.CreateAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描任务记录数据失败: %v", err)
		}
		records = append(records, record)
	}

	// 检查遍历过程中是否有错误
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历任务记录结果集错误: %v", err)
	}

	// 返回查询结果
	return records, nil
}
