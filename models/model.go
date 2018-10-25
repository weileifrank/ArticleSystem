package models

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

type User struct {
	Id       int
	UserName string     `orm:"size(20)"`
	Passwd   string     `orm:"size(20)"`
	IconImg  string     `orm:"default("")"`
	Articles []*Article `orm:"rel(m2m)"`
}

type Article struct {
	Id          int          `orm:"pk;auto"`
	ArtiName    string       `orm:"size(20)"`
	Atime       time.Time    `orm:"auto_now"`
	Acount      int          `orm:"default(0);null"`
	Acontent    string       `orm:"size(500)"`
	Aimg        string       `orm:"size(100)"`
	ArticleType *ArticleType `orm:"rel(fk)"`
	Users       []*User      `orm:"reverse(many)"`
}

type ArticleType struct {
	Id       int
	TypeName string     `orm:"size(20)"`
	Articles []*Article `orm:"reverse(many)"`
}

func init() {
	var dbhost string
	var dbport string
	var dbuser string
	var dbpassword string
	var db string
	//获取配置文件中对应的配置信息
	dbhost = beego.AppConfig.String("dbhost")
	dbport = beego.AppConfig.String("dbport")
	dbuser = beego.AppConfig.String("dbuser")
	dbpassword = beego.AppConfig.String("dbpassword")
	db = beego.AppConfig.String("db")
	orm.RegisterDriver("mysql", orm.DRMySQL) //注册mysql Driver
	//构造conn连接
	conn := dbuser + ":" + dbpassword + "@tcp(" + dbhost + ":" + dbport + ")/" + db + "?charset=utf8"
	//注册数据库连接
	orm.RegisterDataBase("default", "mysql", conn)

	orm.RegisterModel(new(User), new(Article), new(ArticleType)) //注册模型
	orm.RunSyncdb("default", false, true)

}
