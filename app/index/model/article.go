package model

import (
	"github.com/jinzhu/gorm"
	"thinkgin/app/common/model"
	"time"
)

type Article struct {
	model.Model
	TadID      int    `json:"tag_id" gorm:"index"`
	Tag        Tag    `json:"tag"`
	Title      string `json:"title"`
	Desc       string `json:"desc"`
	Content    string `json:"content"`
	CreatedBy  string `json:"created_by"`
	ModifiedBy string `json:"modified_by"`
	State      int    `json:"state"`
}

/**
* 根据ID查询文章是否存在
 */

func ExistArticleByID(id int) bool {
	var article Article
	model.Db.Select("id=?", id).First(&article)
	if article.ID > 0 {
		return true
	}
	return false
}

/**
*文章总数
 */

func GetArticleTotal(maps interface{}) (count int) {
	model.Db.Model(&Article{}).Where(maps).Count(&count)
	return
}

/**
* 获取文章列表
 */

func GetArticles(pageNum int, pageSize int, maps interface{}) (article []Article) {
	model.Db.Preload("Tag").Where(maps).Offset(pageNum).Limit(pageSize).Find(&article)
	return
}

/**
* 获取单个文章
* TODO (article Article) 我写成了 (article *Article)
 */

func GetArticle(id int) (article *Article) {
	model.Db.Where("id=?", id).First(&article)
	model.Db.Model(&article).Related(&article.Tag)
	return
}

/**
* 编辑文章
 */

func EditArticle(id int, data interface{}) bool {
	model.Db.Model(&Article{}).Where("id =?", id).Update(data)
	return true
}

/**
* 新增文章
 */
func AddArticle(data map[string]interface{}) bool {
	model.Db.Create(&Article{
		TadID:     data["tag_id"].(int),
		Title:     data["title"].(string),
		Desc:      data["desc"].(string),
		Content:   data["content"].(string),
		CreatedBy: data["created_by"].(string),
		State:     data["state"].(int),
	})
	return true
}

/**
* 删除文章
* TODO Delete(Article{}) 我用的是  Delete(&Article{})
 */

func DeleteArticle(id int) bool {
	model.Db.Where("id=?", id).Delete(&Article{})
	return true
}

/**
*插入前更新创建时间
 */

func (article *Article) BeforeCreate(scope gorm.Scope) error {
	scope.SetColumn("CreatedOn", time.Now().Unix())
	return nil
}

/**
*更新前写入更新时间
 */

func (article *Article) BeforeUpdate(scope gorm.Scope) error {
	scope.SetColumn("ModifiedOn", time.Now().Unix())
	return nil
}
