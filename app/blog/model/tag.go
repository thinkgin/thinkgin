package model

import (
	"github.com/jinzhu/gorm"
	"thinkgin/app/common/model"
	"time"
)

type Tag struct {
	model.Model

	Name       string `json:"name"`
	CreatedBy  string `json:"created_by"`
	ModifiedBy string `json:"modified_by"`
	State      int    `json:"state"`
}

/*自动写入创建时间*/
func (tag *Tag) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("CreatedOn", time.Now().Unix())
	return nil
}

/*自动写入修改时间*/
func (tag *Tag) BeforeUpdate(scope *gorm.Scope) error {
	scope.SetColumn("ModifiedOn", time.Now().Unix())
	return nil
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

/**
* 编辑标签TAG
 */
func ExistTagById(id int) bool {
	var tag Tag
	model.Db.Select("id").Where("id=?", id).First(&tag)
	if tag.ID > 0 {
		return true
	}
	return false
}

/**
*删除TDG
 */
func DeleteTag(id int) bool {
	var tag Tag
	model.Db.Where("id=?", id).Delete(&tag)
	return true
}

/**
*编辑TDG
 */
func EditTag(id int, data interface{}) bool {
	model.Db.Model(&Tag{}).Where("id=?", id).Updates(data)
	return true
}
