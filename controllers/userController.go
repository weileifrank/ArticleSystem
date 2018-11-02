package controllers

import (
	"ArticleSystem/constants"
	"ArticleSystem/models"
	"encoding/base64"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
)

type UserController struct {
	beego.Controller
}

func (this *UserController) ShowRegister() {
	this.TplName = "register.html"
}

func (this *UserController) ShowLogout() {
	this.DelSession(constants.SESSION_USERNAME)
	this.Redirect(constants.ROOT_URL, 302)
}

func (this *UserController) HandleRegister() {
	userName := this.GetString("userName")
	password := this.GetString("password")
	if userName == "" || password == "" {
		beego.Info("用户名或密码为空")
		this.Data["errmsg"] = "数据不完整"
		this.TplName = "register.html"
		return
	}
	newOrm := orm.NewOrm()
	_, err := newOrm.Insert(&models.User{UserName: userName, Passwd: password})
	if err != nil {
		beego.Info("插入数据错误", err)
		return
	}
	this.Redirect(constants.ROOT_URL, 302)
}

func (this *UserController) ShowLogin() {
	userName := this.Ctx.GetCookie(constants.COOKIE_USERNAME)
	if userName != "" {
		bytes, _ := base64.StdEncoding.DecodeString(userName)
		this.Data["userName"] = string(bytes)
		this.Data["checked"] = "checked"
	} else {
		this.Data["userName"] = ""
		this.Data["checked"] = ""
	}
	this.TplName = "login.html"
}

func (this *UserController) HandleLogin() {
	userName := this.GetString("userName")
	password := this.GetString("password")
	if userName == "" || password == "" {
		beego.Info("用户名或密码为空")
		this.Data["errmsg"] = "数据不完整"
		this.TplName = "login.html"
		return
	}
	newOrm := orm.NewOrm()
	read := newOrm.Read(&models.User{UserName: userName, Passwd: password}, "UserName", "Passwd")
	if read != nil {
		beego.Info("用户名或密码有误,请重新输入", read)
		this.Data["errmsg"] = "用户名或密码有误,请重新输入"
		this.TplName = "login.html"
		return
	}
	remember := this.GetString("remember")
	if remember == "on" {
		username := base64.StdEncoding.EncodeToString([]byte(userName))
		this.Ctx.SetCookie(constants.COOKIE_USERNAME, username, constants.COOKIE_EXPIRE)
	} else {
		this.Ctx.SetCookie(constants.COOKIE_USERNAME, userName, -1)
	}
	this.SetSession(constants.SESSION_USERNAME, userName)
	//conn, e := redis.Dial("tcp", ":6379")
	//if e != nil {
	//	beego.Info("redis.Dial error", e)
	//	return
	//}
	//defer conn.Close()
	//_, err := conn.Do("mset", "key1", "value1", "key2", "value2")
	//if err != nil {
	//	beego.Error("conn.Do error", err)
	//	return
	//}
	this.Redirect(constants.LIST_URL, 302)
}
