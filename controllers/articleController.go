package controllers

import (
	"ArticleSystem/constants"
	"ArticleSystem/models"
	bytes2 "bytes"
	"encoding/gob"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/gomodule/redigo/redis"
	"github.com/weilaihui/fdfs_client"
	"math"
	"path"
	"strconv"
	"time"
)

type ArticleController struct {
	beego.Controller
}

func (this *ArticleController) ShowList() {
	userName := this.GetSession(constants.SESSION_USERNAME)
	if userName == nil {
		this.Redirect(constants.ROOT_URL, 302)
		return
	}

	newOrm := orm.NewOrm()
	querySeter := newOrm.QueryTable("Article")
	var articles []models.Article
	typeName := this.GetString("select")
	var count int64
	pageSize := 2

	pageIndex, err := this.GetInt("pageIndex")
	if err != nil {
		pageIndex = 1
	}
	//起始位置
	start := (pageIndex - 1) * pageSize
	if typeName == "" {
		count, _ = querySeter.Count()
	} else {
		count, _ = querySeter.RelatedSel("ArticleType").Filter("ArticleType__TypeName", typeName).Count()
	}
	pageCount := math.Ceil(float64(count) / float64(pageSize))
	if typeName == "" {
		querySeter.Limit(pageSize, start).All(&articles)

	} else {
		querySeter.Limit(pageSize, start).RelatedSel("ArticleType").Filter("ArticleType__TypeName", typeName).All(&articles)
	}

	//获取文章类型
	var types []models.ArticleType
	dbhost := beego.AppConfig.String("dbhost")
	conn, e := redis.Dial("tcp", dbhost+":6379")
	//strings, i := redis.Strings(conn.Do("mget", "key1", "key2"))
	//if i != nil {
	//	beego.Info("redis.Strings err", i)
	//	return
	//}
	//beego.Info("strings:=======================>", strings)
	if e != nil {
		beego.Error("redis连接错误")
		return
	}
	defer conn.Close()
	reply, err2 := conn.Do("get", "types")
	if reply == nil {
		newOrm.QueryTable("ArticleType").All(&types)
		var buffer bytes2.Buffer
		encoder := gob.NewEncoder(&buffer)
		encode := encoder.Encode(&types)
		if encode != nil {
			beego.Error("encode数据出错")
			return
		}
		_, err3 := conn.Do("set", "types", buffer.Bytes())
		if err3 != nil {
			beego.Error("redis序列化数据出错误")
			return
		}
		beego.Info("从mysql数据库中获取数据了")
	} else {
		//beego.Info("reply:", reply, "=============err2:", err2)
		bytes, i := redis.Bytes(reply, err2)
		if i != nil {
			beego.Error("redis获取数据错误", i)
			return
		}
		decoder := gob.NewDecoder(bytes2.NewBuffer(bytes))
		decode := decoder.Decode(&types)
		if decode != nil {
			beego.Error("decode数据错误")
			return
		}
		beego.Info("types:", types)
		beego.Info("从redis数据库中获取数据了")
	}

	this.Data["types"] = types

	//传递数据
	this.Data["userName"] = userName.(string)
	this.Data["typeName"] = typeName
	this.Data["pageIndex"] = pageIndex
	this.Data["pageCount"] = int(pageCount)
	this.Data["count"] = count
	this.Data["articles"] = articles

	this.Layout = "layout.html"
	this.TplName = "index.html"
}

func (this *ArticleController) ShowAdd() {
	var articleTypes []models.ArticleType
	newOrm := orm.NewOrm()
	_, err := newOrm.QueryTable("ArticleType").All(&articleTypes)
	if err != nil {
		beego.Error("查询出错了", err)
		return
	}
	this.Data["articleTypes"] = articleTypes
	this.Layout = "layout.html"
	this.TplName = "add.html"
}

func (this *ArticleController) HandleAdd() {
	articleName := this.GetString("articleName")
	content := this.GetString("content")
	//2校验数据
	if articleName == "" || content == "" {
		this.Data["errmsg"] = "添加数据不完整"
		this.TplName = "add.html"
		return
	}
	file, header, err := this.GetFile("uploadname")
	defer file.Close()
	if err != nil {
		this.Data["errmsg"] = "文件上传失败"
		this.TplName = "add.html"
		return
	}
	//1.文件大小
	if header.Size > 5000000 {
		this.Data["errmsg"] = "文件太大，请重新上传"
		this.TplName = "add.html"
		return
	}

	//2.文件格式
	ext := path.Ext(header.Filename)
	if ext != ".jpg" && ext != ".png" && ext != ".jpeg" {
		this.Data["errmsg"] = "文件格式错误。请重新上传"
		this.TplName = "add.html"
		return
	}
	//fileName:=time.Now().Format("2006-01-02-15:04:05")+ext
	fileName := strconv.FormatInt(time.Now().UnixNano(), 10)
	//toFile := this.SaveToFile("uploadname", "./static/img/"+fileName)
	//if toFile != nil {
	//	beego.Error("保存文件出错", toFile)
	//	return
	//}
	fileBuffer := make([]byte, header.Size)
	file.Read(fileBuffer)
	client, i := fdfs_client.NewFdfsClient("/etc/fdfs/client.conf")
	if i != nil {
		beego.Info("fdfs_client.NewFdfsClient err:", i)
		return
	}
	response, i2 := client.UploadByBuffer(fileBuffer, ext[1:])
	if i2 != nil {
		beego.Info("client.UploadByBuffer err:", i2)
		return
	}
	beego.Info("response==>", response)
	newOrm := orm.NewOrm()
	//获取类型数据
	typeName := this.GetString("select")
	articleType := models.ArticleType{TypeName: typeName}
	read := newOrm.Read(&articleType, "TypeName")
	if read != nil {
		beego.Error("查询出错了")
		return
	}
	_, e := newOrm.Insert(&models.Article{ArtiName: articleName, Acontent: content, Aimg: "/static/img/" + fileName, ArticleType: &articleType})
	if e != nil {
		beego.Error("插入博文出错", e)
		return
	}
	this.Redirect(constants.LIST_URL, 302)
}

func (this *ArticleController) ShowContent() {
	id, _ := this.GetInt("id")
	newOrm := orm.NewOrm()
	article := models.Article{Id: id}
	read := newOrm.Read(&article)
	if read != nil {
		beego.Error("查询错误", read)
		return
	}
	article.Acount += 1
	newOrm.Update(&article)

	m2M := newOrm.QueryM2M(&article, "Users")
	username := this.GetSession(constants.SESSION_USERNAME)
	if username == nil {
		this.Redirect(constants.ROOT_URL, 302)
		return
	}
	user := models.User{UserName: username.(string)}
	e := newOrm.Read(&user, "UserName")
	if e != nil {
		beego.Error("读取用户出错了", e)
		return
	}
	m2M.Add(user)

	var users []models.User
	newOrm.QueryTable("User").Filter("Articles__Article__Id", id).Distinct().All(&users)
	this.Data["users"] = users
	this.Data["article"] = article
	this.Layout = "layout.html"
	this.TplName = "content.html"
}

func (this *ArticleController) ShowDelete() {
	id, _ := this.GetInt("id")
	newOrm := orm.NewOrm()
	_, err := newOrm.Delete(&models.Article{Id: id})
	if err != nil {
		beego.Error("删除失败")
		return
	}
	this.Redirect(constants.LIST_URL, 302)
}

func (this *ArticleController) ShowUpdate() {
	//获取数据
	id, err := this.GetInt("id")
	//校验数据
	if err != nil {
		beego.Info("请求文章错误")
		return
	}
	//数据处理
	//查询相应文章
	o := orm.NewOrm()
	var article models.Article
	article.Id = id
	o.Read(&article)

	//返回视图
	this.Data["article"] = article
	this.Layout = "layout.html"
	this.TplName = "update.html"
}

func upload(this *beego.Controller, filepath string) string {
	file, header, err := this.GetFile(filepath)
	if err != nil {
		this.Data["errmsg"] = "文件上传失败"
		this.TplName = "add.html"
		return ""
	}
	defer file.Close()
	if header.Filename == "" {
		return "NoImg"
	}
	//1.文件大小
	if header.Size > 5000000 {
		this.Data["errmsg"] = "文件太大，请重新上传"
		this.TplName = "add.html"
		return ""
	}
	//2.文件格式
	ext := path.Ext(header.Filename)
	if ext != ".jpg" && ext != ".png" && ext != ".jpeg" {
		this.Data["errmsg"] = "文件格式错误。请重新上传"
		this.TplName = "add.html"
		return ""
	}
	//3.防止重名
	fileName := strconv.FormatInt(time.Now().UnixNano(), 10) + ext
	//存储
	toFile := this.SaveToFile(filepath, "./static/img/"+fileName)
	if toFile != nil {
		this.Data["errmsg"] = "文件保存出错"
		this.TplName = "add.html"
		return ""
	}
	return "/static/img/" + fileName
}

func (this *ArticleController) HandleUpdate() {
	articleName := this.GetString("articleName")
	content := this.GetString("content")
	if articleName == "" || content == "" {
		beego.Error("博文标题或内容为空")
		return
	}
	id, _ := this.GetInt("id")
	article := models.Article{Id: id}
	newOrm := orm.NewOrm()
	read := newOrm.Read(&article)
	if read != nil {
		beego.Error("更新的博文不存在")
		return
	}
	imgUrl := upload(&this.Controller, "uploadname")
	article.ArtiName = articleName
	article.Acontent = content
	if imgUrl != "NoImg" {
		article.Aimg = imgUrl
	}
	newOrm.Update(&article)
	this.Redirect(constants.LIST_URL, 302)

}

/*=======================================文章类型crud================================================*/

func (this *ArticleController) ShowAddType() {
	newOrm := orm.NewOrm()
	var articleTypes []models.ArticleType
	_, err := newOrm.QueryTable("ArticleType").All(&articleTypes)
	if err != nil {
		beego.Error("查询出错了", err)
		return
	}
	this.Data["articleTypes"] = articleTypes
	this.Layout = "layout.html"
	this.TplName = "addType.html"
}

func (this *ArticleController) HandleAddType() {
	typeName := this.GetString("typeName")
	if typeName == "" {
		beego.Error("文章类型不能为空")
		return
	}
	newOrm := orm.NewOrm()
	_, err := newOrm.Insert(&models.ArticleType{TypeName: typeName})
	if err != nil {
		beego.Error("插入数据失败", err)
		return
	}
	this.Redirect(constants.ADDTYPE_URL, 302)
}

func (this *ArticleController) ShowDeleteType() {
	id, _ := this.GetInt("id")
	newOrm := orm.NewOrm()
	_, err := newOrm.Delete(&models.ArticleType{Id: id})
	if err != nil {
		beego.Error("删除文章类型失败", err)
		return
	}
	this.Redirect(constants.ADDTYPE_URL, 302)
}
