/**
 * @Author: ZhaoYadong
 * @Date: 2023-07-13 17:43:07
 * @LastEditors: ZhaoYadong
 * @LastEditTime: 2023-08-11 17:12:56
 * @FilePath: /shield/util/struct.go
 */
package util

import (
	"time"
)

type ActiveRequest struct {
	Data []byte `json:"data"`
}

type ActiveData struct {
	Code      string `json:"code"`
	MachineID string `json:"machine_id"`
}

type ActiveResponse struct {
	Data []byte `json:"data"`
}
type CDKey struct {
	ProjectName    string    `gorm:"column:project_name;type:varchar(255);not null;"`    // 项目名称
	ProductIDs     string    `gorm:"column:product_ids;type:varchar(255);not null;"`     // 适用产品ID列表
	ProductNames   string    `gorm:"column:product_names;type:varchar(255);not null;"`   // 适用产品名称列表
	ProductSymbols string    `gorm:"column:product_symbols;type:varchar(255);not null;"` // 适用产品标志列表
	Category       int64     `gorm:"column:category;type:tinyint(1);default:1;"`         // 类型;1单品;2集合
	Genre          int64     `gorm:"column:genre;type:tinyint(1);default:1;"`            // 1固定期限,单位天; 2指定时间段有效; 3永久有效
	GenreVal       string    `gorm:"column:genre_val;type:char(50);not null"`            // 类型对应的值
	Status         int64     `gorm:"column:status;type:tinyint(1);default:1;"`           // 状态;由各种时间推断出来;懒加载;获取的时候更新 1待激活;2使用中;3已过期
	Code           string    `gorm:"column:code;type:char(36);not null;unique"`          // 生成的授权码
	MachineID      string    `gorm:"column:machine_id;type:char(50);not null"`           // 设备识别码
	ActivedAt      time.Time `gorm:"column:actived_at;type:datetime;default:null"`       // 激活时间
	StartAt        time.Time `gorm:"column:start_at;type:datetime;default:null"`         // 开始时间
	EndAt          time.Time `gorm:"column:end_at;type:datetime;default:null"`           // 开始时间
	Creator        string    `gorm:"column:creator;type:varchar(50);default:null"`       // 开始时间
}

type Response struct {
	Code int         `json:"code"` // 响应码，0表示成功，非0表示失败
	Msg  string      `json:"msg"`  // 失败消息
	Data interface{} `json:"data"` // 通用数据返回字段
}

type Encrypt struct {
	CDkey     string //激活码
	MachineID string //设备识别码

	ActiveAt time.Time //激活时间
	SyncedAt time.Time //同步时钟
	StartAt  time.Time //软件开始时间
	EndAt    time.Time //软件过期时间
	Genre    int64     //1固定期限,单位天; 2指定时间段有效; 3永久有效

	Scope map[string]bool //适用软件范围,默认都不适用,key 软件标识
}

type Runable interface {
	Exec()
}
