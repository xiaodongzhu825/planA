package sqLite

import (
	"database/sql"
	"errors"
	"fmt"
	"planA/initialization/golabl"
	"planA/tool"
	sqLiteType "planA/type/sqLite"
	"strings"
	"time"
)

// CreateTaskExportTab 创建task_export表
// @return error 错误信息
func CreateTaskExportTab() error {
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS task_export (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id VARCHAR(100) NOT NULL,
        shop_id VARCHAR(100) NOT NULL,
        task_id VARCHAR(100) NOT NULL,
        shop_name VARCHAR(100) NOT NULL,
        file_url VARCHAR(300),
        status INTEGER NOT NULL DEFAULT 0,
        total INTEGER NOT NULL DEFAULT 0,
        complete_at DATETIME,
        create_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
    );

    CREATE INDEX IF NOT EXISTS idx_task_id ON task_export(task_id);
    CREATE INDEX IF NOT EXISTS idx_status ON task_export(status);
    CREATE INDEX IF NOT EXISTS idx_create_at ON task_export(create_at);
`
	_, err := golabl.SqliteDb.Exec(createTableSQL)
	if err != nil {
		return err
	}
	return nil
}

// CreateTaskExport 向task_export表插入一条记录
// @param export TaskExport 要插入的导出记录
// @return int64 插入记录的自增ID
// @return error 错误信息
func CreateTaskExport(export sqLiteType.TaskExport) (int64, error) {
	// 在 Go 代码中计算当前时间
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	insertSQL := `INSERT INTO task_export (user_id, shop_id, task_id, shop_name, file_url, status,total,create_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := golabl.SqliteDb.Exec(
		insertSQL,
		export.UserID,   // user_id
		export.ShopID,   // shop_id
		export.TaskID,   // task_id
		export.ShopName, // shop_name
		export.FileUrl,  // file_url（允许空字符串）
		export.Status,   // status
		export.Total,    // total
		currentTime,
	)
	if err != nil {
		return 0, err
	}
	// 获取插入记录的自增ID（SQLite中LastInsertId()返回rowid）
	insertID, err := result.LastInsertId()
	if err != nil {
		return 0, err // 获取ID失败返回0和错误
	}

	return insertID, nil // 成功返回自增ID和nil
}

// UpdateTaskExportStatus 更新task_export表中的status字段
// @param taskId string 任务Id
// @param status int64 状态
// @param fullPath string 文件路径
// @return error 错误信息
func UpdateTaskExportStatus(taskId string, status int64, fullPath string) error {
	var err error
	if status == 2 {
		// 当status=2时，同时更新status、complete_at和file_url字段
		_, err = golabl.SqliteDb.Exec(
			"UPDATE task_export SET status = ?, complete_at = ?, file_url = ? WHERE task_id = ?",
			status,
			time.Now().Format("2006-01-02 15:04:05"), // 设置为当前系统时间
			fullPath,
			taskId,
		)
	} else {
		// 其他状态只更新status字段
		_, err = golabl.SqliteDb.Exec(
			"UPDATE task_export SET status = ? WHERE task_id = ?",
			status,
			taskId,
		)
	}
	return err
}

// GetTaskExportsList 分页查询task_export表记录（无查询条件）
// @param page 页码（从1开始）
// @param pageSize 每页条数
// @param userId 用户ID
// @return []TaskExport 记录列表
// @return int64 总记录数
// @return error 错误信息
func GetTaskExportsList(page, pageSize int, userId string) ([]sqLiteType.TaskExport, int64, error) {
	// 参数校验
	pageSize, offset := tool.GetPage(page, pageSize)

	// 构建查询条件（当前为空）
	var conditions []string
	var args []interface{}

	// 如果userId不为空，则添加用户ID条件
	if userId != "" {
		conditions = append(conditions, "user_id = ?")
		args = append(args, userId)
	}

	// 构建 WHERE子句
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 查询总数
	var total int64
	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM task_export %s`, whereClause)

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
        SELECT id, user_id, task_id, shop_name, file_url, status, total, complete_at, create_at 
        FROM task_export 
        %s
        ORDER BY create_at DESC 
        LIMIT ? OFFSET ?
    `, whereClause)
	// 添加分页参数到args
	queryArgs := append(args, pageSize, offset)

	rows, err := golabl.SqliteDb.Query(querySQL, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询失败: %v", err)
	}
	defer rows.Close()

	var records []sqLiteType.TaskExport

	for rows.Next() {
		var record sqLiteType.TaskExport
		// 扫描所有字段
		err = rows.Scan(
			&record.ID,
			&record.UserID,
			&record.TaskID,
			&record.ShopName,
			&record.FileUrl,
			&record.Status,
			&record.Total,
			&record.CompleteAt,
			&record.CreateAt,
		)
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

// DeleteOldExport 删除task_export表中3天前的记录
// @return error 错误信息
func DeleteOldExport() error {
	// 使用SQLite的date函数计算3天前
	result, err := golabl.SqliteDb.Exec(`
        DELETE FROM task_export 
        WHERE create_at < datetime('now','localtime', '-3 days')
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

// GetOldExport 获取task_export表中3天前的记录
// @return []TaskExport 3天前的记录
// @return error 错误信息
func GetOldExport() ([]sqLiteType.TaskExport, error) {
	// 计算3天前的时间
	sevenDaysAgo := time.Now().AddDate(0, 0, -3)

	// 查询3天前的记录
	rows, err := golabl.SqliteDb.Query(`
        SELECT id, user_id, task_id, shop_name, file_url, status, total, complete_at, create_at 
        FROM task_export 
        WHERE create_at < ? 
        ORDER BY create_at ASC
    `, sevenDaysAgo)

	if err != nil {
		return nil, fmt.Errorf("查询3天前导出记录失败: %v", err)
	}
	defer rows.Close()

	var tasks []sqLiteType.TaskExport

	for rows.Next() {
		var task sqLiteType.TaskExport
		var completeAt, createAt sql.NullTime

		err := rows.Scan(
			&task.ID,
			&task.UserID,
			&task.TaskID,
			&task.ShopName,
			&task.FileUrl,
			&task.Status,
			&task.Total,
			&completeAt,
			&createAt,
		)

		if err != nil {
			return nil, fmt.Errorf("扫描导出记录失败: %v", err)
		}

		// 转换时间字段
		if completeAt.Valid {
			task.CompleteAt = sql.NullTime{
				Time:  completeAt.Time,
				Valid: true,
			}
		}
		if createAt.Valid {
			task.CreateAt = createAt.Time
		}

		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历导出记录失败: %v", err)
	}

	return tasks, nil
}

// GetTaskExportByTaskID 根据任务ID获取导出记录
// @param taskID string 任务ID
// @return sqLiteType.TaskExport 导出记录
// @return error 错误信息
func GetTaskExportByTaskID(taskID string) (sqLiteType.TaskExport, error) {
	query := `SELECT id, user_id, shop_id, task_id, shop_name, file_url, 
                     status, total, complete_at, create_at 
              FROM task_export 
              WHERE task_id = ?`

	var task sqLiteType.TaskExport
	err := golabl.SqliteDb.QueryRow(query, taskID).Scan(
		&task.ID,
		&task.UserID,
		&task.ShopID,
		&task.TaskID,
		&task.ShopName,
		&task.FileUrl,
		&task.Status,
		&task.Total,
		&task.CompleteAt,
		&task.CreateAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return task, nil // 未找到记录
		}
		return task, err
	}

	return task, nil
}

// UpdateTaskExport 更新task_export信息
// @param export sqLiteType.TaskExport 要更新的任务信息
// @return error 错误信息
func UpdateTaskExport(export sqLiteType.TaskExport) error {
	query := `
        UPDATE task_export 
        SET user_id = ?,
            shop_id = ?,
            task_id = ?,
            shop_name = ?,
            file_url = ?,
            status = ?,  
            total = ?,
            complete_at = ?
        WHERE task_id = ?
    `

	result, err := golabl.SqliteDb.Exec(query, export.TaskID, export.ShopID, export.TaskID, export.ShopName, export.FileUrl, export.Status, export.Total, export.CompleteAt, export.TaskID)
	if err != nil {
		return fmt.Errorf("更新任务失败: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("未找到task_id为 %s 的任务", export.TaskID)
	}
	return nil
}
