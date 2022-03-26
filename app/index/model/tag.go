package model

import "thinkgin/app/common/model"

type Tag struct {
	model.Model

	Name       string `json:"name"`
	CreatedBy  string `json:"created_by"`
	ModifiedBy string `json:"modified_by"`
	State      int    `json:"state"`
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

/**
* 检测tag是否存在
 */
func ExistTagByName(name string) bool {
	var tag Tag
	model.Db.Select("id").Where("name=?", name).First(&tag)
	if tag.ID > 0 {
		return true
	}
	return false
}

/**
* 新增标签tag
 */
func AddTag(name string, state int, createdBy string) bool {
	model.Db.Create(&Tag{
		Name:      name,
		State:     state,
		CreatedBy: createdBy,
	})
	return true
}
