## /thinkgin-new
```text
再来一遍
```
#### 公共Header参数
参数名 | 示例值 | 参数类型 | 是否必填 | 参数描述
--- | --- | --- | --- | ---
暂无参数
#### 公共Query参数
参数名 | 示例值 | 参数类型 | 是否必填 | 参数描述
--- | --- | --- | --- | ---
暂无参数
#### 公共Body参数
参数名 | 示例值 | 参数类型 | 是否必填 | 参数描述
--- | --- | --- | --- | ---
暂无参数
#### 预执行脚本
```javascript
暂无预执行脚本
```
#### 后执行脚本
```javascript
暂无后执行脚本
```
## /thinkgin-new/测试接口
```text
暂无描述
```
#### 接口状态
> 已完成

#### 接口URL
> 127.0.0.1:8000/test

#### 请求方式
> GET

#### Content-Type
> form-data

#### 预执行脚本
```javascript
暂无预执行脚本
```
#### 后执行脚本
```javascript
暂无后执行脚本
```
## /thinkgin-new/获取标签列表
```text
暂无描述
```
#### 接口状态
> 已完成

#### 接口URL
> 127.0.0.1:8000/api/v1/tags?page=2&name=&state=

#### 请求方式
> GET

#### Content-Type
> form-data

#### 请求Query参数
参数名 | 示例值 | 参数类型 | 是否必填 | 参数描述
--- | --- | --- | --- | ---
page | 2 | Number | 否 | -
name | - | String | 否 | 标签名称查询指定标签信息
state | - | Number | 否 | 状态 0为禁用、1为启用
#### 预执行脚本
```javascript
暂无预执行脚本
```
#### 后执行脚本
```javascript
暂无后执行脚本
```
#### 成功响应示例
```javascript
{
	"code": 200,
	"data": {
		"lists": [
			{
				"id": 1,
				"created_on": 1648048265,
				"modified_on": 0,
				"name": "news123",
				"created_by": "laowu",
				"modified_by": "",
				"state": 1
			}
		],
		"total": 4
	},
	"msg": "ok"
}
```
参数名 | 示例值 | 参数类型 | 是否必填 | 参数描述
--- | --- | --- | --- | ---
code | 200 | Number | 是 | 
data | - | Object | 是 | 返回数据
data.lists | - | Object | 是 | 
data.lists.id | 1 | Number | 是 | ID
data.lists.created_on | 1648048265 | Number | 是 | 创建时间
data.lists.modified_on | - | Number | 是 | 修改时间
data.lists.name | news123 | String | 是 | 标签名
data.lists.created_by | laowu | String | 是 | 创建人
data.lists.modified_by | - | Object | 是 | 修改人
data.lists.state | 1 | Number | 是 | 状态 0为禁用、1为启用
data.total | 4 | Number | 是 | 总数
msg | ok | String | 是 | 返回文字描述
## /thinkgin-new/新增标签
```text
暂无描述
```
#### 接口状态
> 开发中

#### 接口URL
> 127.0.0.1:8000/api/v1/tags

#### 请求方式
> POST

#### Content-Type
> form-data

#### 请求Body参数
参数名 | 示例值 | 参数类型 | 是否必填 | 参数描述
--- | --- | --- | --- | ---
name | goods | Text | 是 | 标签名
state | 1 | Text | 是 | 状态 0为禁用、1为启用
created_by | test | Text | 是 | 创建人
#### 预执行脚本
```javascript
暂无预执行脚本
```
#### 后执行脚本
```javascript
暂无后执行脚本
```
#### 成功响应示例
```javascript
{
	"code": 200,
	"data": {},
	"message": "ok"
}
```
参数名 | 示例值 | 参数类型 | 是否必填 | 参数描述
--- | --- | --- | --- | ---
code | 200 | Number | 是 | 
data | {} | Object | 是 | 返回数据
message | ok | String | 是 | 
#### 错误响应示例
```javascript
{
	"code": 10001,
	"data": {},
	"message": "已存在该标签名称"
}
```
参数名 | 示例值 | 参数类型 | 是否必填 | 参数描述
--- | --- | --- | --- | ---
code | 10001 | Number | 是 | 
data | {} | Object | 是 | 返回数据
message | 已存在该标签名称 | String | 是 | 
## /thinkgin-new/修改标签
```text
暂无描述
```
#### 接口状态
> 开发中

#### 接口URL
> 127.0.0.1:8000/api/v1/tags/3

#### 请求方式
> PUT

#### Content-Type
> form-data

#### 请求Query参数
参数名 | 示例值 | 参数类型 | 是否必填 | 参数描述
--- | --- | --- | --- | ---
name | goods6 | Text | 是 | -
modified_by | test12 | Text | 是 | -
state | 1 | Text | 是 | -
#### 请求Body参数
参数名 | 示例值 | 参数类型 | 是否必填 | 参数描述
--- | --- | --- | --- | ---
name | goods33333 | String | 是 | 标签名
modified_by | test12 | String | 是 | 修改人
state | 1 | Number | 是 | 状态 0为禁用、1为启用
#### 预执行脚本
```javascript
暂无预执行脚本
```
#### 后执行脚本
```javascript
暂无后执行脚本
```
#### 成功响应示例
```javascript
{
	"code": 200,
	"data": {},
	"message": "ok"
}
```
参数名 | 示例值 | 参数类型 | 是否必填 | 参数描述
--- | --- | --- | --- | ---
code | 200 | Number | 是 | 
data | {} | Object | 是 | 返回数据
message | ok | String | 是 | 
#### 错误响应示例
```javascript
{
	"code": 10001,
	"data": {},
	"message": "已存在该标签名称"
}
```
参数名 | 示例值 | 参数类型 | 是否必填 | 参数描述
--- | --- | --- | --- | ---
code | 10001 | Number | 是 | 
data | {} | Object | 是 | 返回数据
message | 已存在该标签名称 | String | 是 | 
## /thinkgin-new/删除标签
```text
暂无描述
```
#### 接口状态
> 开发中

#### 接口URL
> 127.0.0.1:8000/api/v1/tags/81

#### 请求方式
> DELETE

#### Content-Type
> form-data

#### 请求Body参数
参数名 | 示例值 | 参数类型 | 是否必填 | 参数描述
--- | --- | --- | --- | ---
name | goods | Text | 是 | 标签名
state | 1 | Text | 是 | 状态 0为禁用、1为启用
created_by | test | Text | 是 | 创建人
#### 预执行脚本
```javascript
暂无预执行脚本
```
#### 后执行脚本
```javascript
暂无后执行脚本
```
#### 成功响应示例
```javascript
{
	"code": 200,
	"data": {},
	"message": "ok"
}
```
参数名 | 示例值 | 参数类型 | 是否必填 | 参数描述
--- | --- | --- | --- | ---
code | 200 | Number | 是 | 
data | {} | Object | 是 | 返回数据
message | ok | String | 是 | 
#### 错误响应示例
```javascript
{
	"code": 10001,
	"data": {},
	"message": "已存在该标签名称"
}
```
参数名 | 示例值 | 参数类型 | 是否必填 | 参数描述
--- | --- | --- | --- | ---
code | 10001 | Number | 是 | 
data | {} | Object | 是 | 返回数据
message | 已存在该标签名称 | String | 是 | 
## /thinkgin-new/修改文章
```text
暂无描述
```
#### 接口状态
> 开发中

#### 接口URL
> 127.0.0.1:8000/api/v1/articles/3

#### 请求方式
> PUT

#### Content-Type
> form-data

#### 请求Body参数
参数名 | 示例值 | 参数类型 | 是否必填 | 参数描述
--- | --- | --- | --- | ---
tag_id | 3 | Number | 是 | 标签ID
title | 文章1 | Text | 是 | 创建人
desc | 这文章 | Text | 是 | 文章描述
content | 哈哈哈哈 | Text | 是 | 文章内容
created_by | 老吴 | Text | 是 | 创建人
state | 1 | Text | 是 | 状态 0为禁用、1为启用
modified_by | 老王 | Text | 是 | 修改人
#### 预执行脚本
```javascript
暂无预执行脚本
```
#### 后执行脚本
```javascript
暂无后执行脚本
```
#### 成功响应示例
```javascript
{
	"code": 200,
	"data": {},
	"message": "ok"
}
```
参数名 | 示例值 | 参数类型 | 是否必填 | 参数描述
--- | --- | --- | --- | ---
code | 200 | Number | 是 | 
data | {} | Object | 是 | 返回数据
message | ok | String | 是 | 
#### 错误响应示例
```javascript
{
	"code": 10001,
	"data": {},
	"message": "已存在该标签名称"
}
```
参数名 | 示例值 | 参数类型 | 是否必填 | 参数描述
--- | --- | --- | --- | ---
code | 10001 | Number | 是 | 
data | {} | Object | 是 | 返回数据
message | 已存在该标签名称 | String | 是 | 
## /thinkgin-new/查询文章
```text
暂无描述
```
#### 接口状态
> 开发中

#### 接口URL
> 127.0.0.1:8000/api/v1/articles/1

#### 请求方式
> GET

#### Content-Type
> form-data

#### 预执行脚本
```javascript
暂无预执行脚本
```
#### 后执行脚本
```javascript
暂无后执行脚本
```
#### 成功响应示例
```javascript
{
	"code": 200,
	"data": {},
	"message": "ok"
}
```
参数名 | 示例值 | 参数类型 | 是否必填 | 参数描述
--- | --- | --- | --- | ---
code | 200 | Number | 是 | 
data | {} | Object | 是 | 返回数据
message | ok | String | 是 | 
#### 错误响应示例
```javascript
{
	"code": 10001,
	"data": {},
	"message": "已存在该标签名称"
}
```
参数名 | 示例值 | 参数类型 | 是否必填 | 参数描述
--- | --- | --- | --- | ---
code | 10001 | Number | 是 | 
data | {} | Object | 是 | 返回数据
message | 已存在该标签名称 | String | 是 | 
## /thinkgin-new/查询文章列表
```text
暂无描述
```
#### 接口状态
> 开发中

#### 接口URL
> 127.0.0.1:8000/api/v1/articles?page=2

#### 请求方式
> GET

#### Content-Type
> form-data

#### 请求Query参数
参数名 | 示例值 | 参数类型 | 是否必填 | 参数描述
--- | --- | --- | --- | ---
page | 2 | Text | 是 | -
#### 请求Body参数
参数名 | 示例值 | 参数类型 | 是否必填 | 参数描述
--- | --- | --- | --- | ---
state | 1 | Text | 是 | 状态 0为禁用、1为启用
#### 预执行脚本
```javascript
暂无预执行脚本
```
#### 后执行脚本
```javascript
暂无后执行脚本
```
#### 成功响应示例
```javascript
{
	"code": 200,
	"data": {},
	"message": "ok"
}
```
参数名 | 示例值 | 参数类型 | 是否必填 | 参数描述
--- | --- | --- | --- | ---
code | 200 | Number | 是 | 
data | {} | Object | 是 | 返回数据
message | ok | String | 是 | 
#### 错误响应示例
```javascript
{
	"code": 10001,
	"data": {},
	"message": "已存在该标签名称"
}
```
参数名 | 示例值 | 参数类型 | 是否必填 | 参数描述
--- | --- | --- | --- | ---
code | 10001 | Number | 是 | 
data | {} | Object | 是 | 返回数据
message | 已存在该标签名称 | String | 是 | 
## /thinkgin-new/新增文章
```text
暂无描述
```
#### 接口状态
> 开发中

#### 接口URL
> 127.0.0.1:8000/api/v1/articles

#### 请求方式
> POST

#### Content-Type
> form-data

#### 请求Body参数
参数名 | 示例值 | 参数类型 | 是否必填 | 参数描述
--- | --- | --- | --- | ---
tag_id | 3 | Number | 是 | 标签ID
title | 文章1 | Text | 是 | 创建人
desc | 这文章 | Text | 是 | 文章描述
content | 哈哈哈哈 | Text | 是 | 文章内容
created_by | 老吴 | Text | 是 | 创建人
state | 1 | Text | 是 | 状态 0为禁用、1为启用
#### 预执行脚本
```javascript
暂无预执行脚本
```
#### 后执行脚本
```javascript
暂无后执行脚本
```
#### 成功响应示例
```javascript
{
	"code": 200,
	"data": {},
	"message": "ok"
}
```
参数名 | 示例值 | 参数类型 | 是否必填 | 参数描述
--- | --- | --- | --- | ---
code | 200 | Number | 是 | 
data | {} | Object | 是 | 返回数据
message | ok | String | 是 | 
#### 错误响应示例
```javascript
{
	"code": 10001,
	"data": {},
	"message": "已存在该标签名称"
}
```
参数名 | 示例值 | 参数类型 | 是否必填 | 参数描述
--- | --- | --- | --- | ---
code | 10001 | Number | 是 | 
data | {} | Object | 是 | 返回数据
message | 已存在该标签名称 | String | 是 | 
## /thinkgin-new/删除文章
```text
暂无描述
```
#### 接口状态
> 已完成

#### 接口URL
> 127.0.0.1:8000/api/v1/articles/4

#### 请求方式
> DELETE

#### Content-Type
> form-data

#### 预执行脚本
```javascript
暂无预执行脚本
```
#### 后执行脚本
```javascript
暂无后执行脚本
```
#### 成功响应示例
```javascript
{
	"code": 200,
	"data": {},
	"message": "ok"
}
```
参数名 | 示例值 | 参数类型 | 是否必填 | 参数描述
--- | --- | --- | --- | ---
code | 200 | Number | 是 | 
data | {} | Object | 是 | 返回数据
message | ok | String | 是 | 
#### 错误响应示例
```javascript
{
	"code": 10001,
	"data": {},
	"message": "已存在该标签名称"
}
```
参数名 | 示例值 | 参数类型 | 是否必填 | 参数描述
--- | --- | --- | --- | ---
code | 10001 | Number | 是 | 
data | {} | Object | 是 | 返回数据
message | 已存在该标签名称 | String | 是 | 
