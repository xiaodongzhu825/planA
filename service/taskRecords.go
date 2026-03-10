package service

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
        user_id INTEGER NOT NULL DEFAULT 0,
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

// InsertTaskRecord 向task_records表插入一条记录
// @param record TaskRecord 要插入的记录
// @return error 错误信息
func InsertTaskRecord(record sqLiteType.TaskRecord) error {
	// 在 Go 代码中计算当前时间
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	insertSQL := `INSERT INTO task_records (user_id,task_id,shop_name,task_type, create_at) VALUES (?, ?, ?, ?, ?)`
	result, err := golabl.SqliteDb.Exec(insertSQL, record.UserID, record.TaskID, record.ShopName, record.TaskType, currentTime)
	if err != nil {
		return err
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	record.ID = int(lastID)
	return nil
}

// GetTaskRecordsWithPage 分页查询task_records表记录
// @param page 页码（从1开始）
// @param pageSize 每页条数
// @param taskId 任务ID（可选，为空时不作为条件）
// @param shopName 店铺名称（可选，为空时不作为条件）
// @param taskType 任务类型（可选，为空时不作为条件）
// @return []TaskRecord 记录列表
// @return int64 总记录数
// @return error 错误信息
func GetTaskRecordsWithPage(page, pageSize int, taskId string, shopName string, taskType int) ([]sqLiteType.TaskRecord, int64, error) {
	// 参数校验
	pageSize, offset := tool.GetPage(page, pageSize)

	// 构建查询条件
	var conditions []string
	var args []interface{}

	if taskId != "" {
		conditions = append(conditions, "task_id = ?")
		args = append(args, taskId)
	}

	if shopName != "" {
		conditions = append(conditions, "shop_name = ?")
		args = append(args, shopName)
	}
	if taskType != 0 {
		conditions = append(conditions, "task_type = ?")
		args = append(args, taskType)
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
        SELECT id,user_id, task_id, shop_name,is_export, task_type,create_at 
        FROM task_records 
        %s
        ORDER BY create_at DESC 
        LIMIT ? OFFSET ?
    `, whereClause)

	// 添加分页参数到 args
	queryArgs := append(args, pageSize, offset)

	rows, err := golabl.SqliteDb.Query(querySQL, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询失败: %v", err)
	}
	defer rows.Close()

	var records []sqLiteType.TaskRecord

	for rows.Next() {
		var record sqLiteType.TaskRecord
		err = rows.Scan(&record.ID, &record.UserID, &record.TaskID, &record.ShopName, &record.IsExport, &record.TaskType, &record.CreateAt)
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

// GetTaskRecordByTaskID 根据taskId查询单个任务记录
// @param taskID 任务ID
// @return *TaskRecord 记录指针
// @return error 错误信息
func GetTaskRecordByTaskID(taskID string) (*sqLiteType.TaskRecord, error) {
	query := `SELECT id,user_id, task_id, shop_name, task_type, create_at 
              FROM task_records 
              WHERE task_id = ? 
              LIMIT 1`

	var record sqLiteType.TaskRecord
	var createAtStr string

	err := golabl.SqliteDb.QueryRow(query, taskID).Scan(
		&record.ID,
		&record.UserID,
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

// GetTaskRecordsByTaskID 根据任务id查询任务记录（task_records表）
// 注意：函数名可能重复，建议重命名，这里保持原样
// @param taskID 任务id
// @return *TaskRecord 任务记录
// @return error 错误信息
func GetTaskRecordsByTaskID(taskID string) (*sqLiteType.TaskRecord, error) {
	query := `SELECT id, task_id, shop_name, is_export, task_type, create_at 
              FROM task_records
              WHERE task_id = ? 
              LIMIT 1`
	var export sqLiteType.TaskRecord
	err := golabl.SqliteDb.QueryRow(query, taskID).Scan(
		&export.ID,
		&export.TaskID,
		&export.ShopName,
		&export.IsExport,
		&export.TaskType,
		&export.CreateAt,
	)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %v", err)
	}
	return &export, nil
}

// UpdateTaskRecordIsExport 更新task_records表的IsExport字段为1
// @param taskID 任务id
// @return error 错误信息
func UpdateTaskRecordIsExport(taskID string) error {
	_, err := golabl.SqliteDb.Exec("UPDATE task_records SET is_export = 1 WHERE task_id = ?", taskID)
	return err
}

// GetTaskRecordById 根据id查询task_records表id=1的数据（仅用于测试心跳）
func GetTaskRecordById() {
	var record sqLiteType.TaskRecord
	golabl.SqliteDb.QueryRow("SELECT * FROM task_records WHERE id = ?", 1).Scan(&record.ID, &record.TaskID, &record.ShopName, &record.CreateAt)
}

// DeleteOldRecordsSQLite 删除task_records表中7天前的记录
// @return error 错误信息
func DeleteOldRecordsSQLite() error {
	// 使用SQLite的date函数计算7天前
	result, err := golabl.SqliteDb.Exec(`
        DELETE FROM task_records 
        WHERE create_at < datetime('now', '-7 days')
    `)
	if err != nil {
		return fmt.Errorf("删除旧数据失败: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %v", err)
	}

	fmt.Printf("已删除 %d 条大于7天的记录\n", rowsAffected)
	return nil
}
