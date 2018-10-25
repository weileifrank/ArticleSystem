package routers

import (
	"ArticleSystem/constants"
	"ArticleSystem/controllers"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

func init() {

	beego.InsertFilter("/article/*", beego.BeforeExec, FilterArticleUrl)
	//user中的注册和登录
	beego.Router(constants.ROOT_URL, &controllers.UserController{}, "get:ShowLogin;post:HandleLogin")
	beego.Router(constants.REGISTER_URL, &controllers.UserController{}, "get:ShowRegister;post:HandleRegister")
	beego.Router(constants.LOGOUT_URL, &controllers.UserController{}, "get:ShowLogout")

	//article中的crud
	beego.Router(constants.LIST_URL, &controllers.ArticleController{}, "get:ShowList")
	beego.Router(constants.Add_URL, &controllers.ArticleController{}, "get:ShowAdd;post:HandleAdd")
	beego.Router(constants.DELETE_URL, &controllers.ArticleController{}, "get:ShowDelete")
	beego.Router(constants.CONTENT_URL, &controllers.ArticleController{}, "get:ShowContent")
	beego.Router(constants.UPDATE_URL, &controllers.ArticleController{}, "get:ShowUpdate;post:HandleUpdate")

	//article_type crud
	beego.Router(constants.ADDTYPE_URL, &controllers.ArticleController{}, "get:ShowAddType;post:HandleAddType")
	beego.Router(constants.DELETETYPE_URL, &controllers.ArticleController{}, "get:ShowDeleteType")
}
func FilterArticleUrl(ctx *context.Context) {
	username := ctx.Input.Session(constants.SESSION_USERNAME)
	if username == nil {
		beego.Info("被拦截了	")
		ctx.Redirect(302, constants.ROOT_URL)
		return
	}
}
