package common

import "github.com/shadow1ng/fscan/client"

type ContextHolder struct {
	Redis *client.RedisContext
	Minio *client.MinioContext
	Mysql *client.MysqlContext
}

type ScanTask struct {
	TaskId    string
	RecordId  string
	StartTime int64
	EndTime   int64
	Status    TaskStatus
}

// TaskStatus 扫描任务执行状态
type TaskStatus string

const (
	//BEGINNING TaskStatus = "BEGINNING" // 已开始
	SCANNING TaskStatus = "SCANNING" // 扫描中
	DONE     TaskStatus = "DONE"     // 已结束
)

// ScanTaskHolder 保存了当前运行中的所有扫描任务信息。key: RecordId    value: ScanTask
var ScanTaskHolder = make(map[string]*ScanTask)

var Context = ContextHolder{
	Redis: &client.RedisContext{},
	Minio: &client.MinioContext{},
	Mysql: &client.MysqlContext{},
}
