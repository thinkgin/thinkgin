package model

import "thinkgin/app/common/model"

type Tag struct {
	model.Model

	Name       string `json:"name"`
	CreateBy   string `json:"created_by"`
	ModifiedBy string `json:"modified_by"`
	State      string `json:"state"`
}

/**
 * 获取分类列表
 */
func GetTags(pageNum int, pageSize int, maps interface{}) (tags []Tag) {
	model.Db.Where(maps).Offset(pageNum).Limit(pageSize).Find(&tags)
	return
}

/**
* 获取Tag总数
 */
func GetTagTotal(maps interface{}) (count int) {
	model.Db.Model(&Tag{}).Where(maps).Count(&count)
	return
}
