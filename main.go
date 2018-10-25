package main

import (
	_ "ArticleSystem/models"
	_ "ArticleSystem/routers"
	"github.com/astaxie/beego"
)

func main() {
	beego.AddFuncMap("prepage", ShowPrePage)
	beego.AddFuncMap("nextpage", ShowNextPage)
	beego.Run()
}

func ShowPrePage(pageIndex int) int {
	if pageIndex == 1 {
		return pageIndex
	}
	return pageIndex - 1
}

func ShowNextPage(pageIndex, pageCount int) int {
	if pageIndex == pageCount {
		return pageIndex
	}
	return pageIndex + 1
}
