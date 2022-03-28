package v1

import (
	"github.com/astaxie/beego/validation"
	"github.com/gin-gonic/gin"
	"github.com/unknwon/com"
	"net/http"
	"thinkgin/app/index/model"
	error2 "thinkgin/extend/error"
	"thinkgin/extend/setting"
	"thinkgin/extend/util"
)

//获取多个文章标签
func GetTags(c *gin.Context) {
	name := c.Query("name")
	maps := make(map[string]interface{})
	data := make(map[string]interface{})
	if name != "" {
		maps["name"] = name
	}

	var state int = -1
	if arg := c.Query("state"); arg != "" {
		state = com.StrTo(arg).MustInt()
		maps["state"] = state
	}

	code := error2.SUCCESS
	data["total"] = model.GetTagTotal(maps)
	data["lists"] = model.GetTags(util.GetPage(c), setting.PageSize, maps)

	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  error2.GetMsg(code),
		"data": data,
	})
}

//新增文章标签
func AddTag(c *gin.Context) {
	name := c.PostForm("name")
	state := com.StrTo(c.DefaultPostForm("state", "0")).MustInt()
	createdBy := c.PostForm("created_by")

	valid := validation.Validation{}
	valid.Required(name, "name").Message("名称不能为空")
	valid.MaxSize(name, 100, "name").Message("名称最大为100个字符")
	valid.Required(createdBy, "created_by").Message("创建人不能为空")
	valid.MaxSize(createdBy, 100, "created_by").Message("创建人最大为100个字符")
	valid.Range(state, 0, 1, "state").Message("状态只允许0或1")

	code := error2.INVALID_PARAMS
	if !valid.HasErrors() {
		if !model.ExistTagByName(name) {
			code = error2.SUCCESS
			model.AddTag(name, state, createdBy)
		} else {
			code = error2.ERROR_EXIST_TAG
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    code,
		"message": error2.GetMsg(code),
		"data":    make(map[string]string),
	})
}

//修改文章标签
func EditTag(c *gin.Context) {
	id := com.StrTo(c.Param("id")).MustInt()
	name := c.PostForm("name")
	modifiedBy := c.PostForm("modified_by")

	valid := validation.Validation{}

	var state int = -1
	if arg := c.PostForm("state"); arg != "" {
		state = com.StrTo(arg).MustInt()
		valid.Range(state, 0, 1, "state").Message("状态只允许0或1")
	}

	valid.Required(id, "id").Message("ID不能为空")
	valid.Required(modifiedBy, "modified_by").Message("修改人不能为空")
	valid.MaxSize(modifiedBy, 100, "modified_by").Message("修改人最长为100字符")
	valid.MaxSize(name, 100, "name").Message("名称最长为100字符")
	code := error2.INVALID_PARAMS

	if !valid.HasErrors() {
		code = error2.SUCCESS
		if model.ExistTagById(id) {
			data := make(map[string]interface{})
			data["modified_by"] = modifiedBy
			if name != "" {
				data["name"] = name
			}
			if state != -1 {
				data["state"] = state
			}
			model.EditTag(id, data)
		} else {
			code = error2.ERROR_NOT_EXIST_TAG
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  error2.GetMsg(code),
		"data": make(map[string]interface{}),
	})

}

//删除文章标签
func DelTag(c *gin.Context) {
	id := com.StrTo(c.Param("id")).MustInt()
	valid := validation.Validation{}
	valid.Min(id, 0, "id").Message("ID必须大于零")

	code := error2.INVALID_PARAMS
	if !valid.HasErrors() {
		code = error2.SUCCESS
		if model.ExistTagById(id) {
			model.DeleteTag(id)
		} else {
			code = error2.ERROR_NOT_EXIST_TAG
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  error2.GetMsg(code),
		"data": make(map[string]interface{}),
	})
}
