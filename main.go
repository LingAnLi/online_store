package main

import (
	_ "GoodsShop/routers"
	"github.com/astaxie/beego"
)

func main() {
	beego.AddFuncMap("Pageup",Pageup)
	beego.AddFuncMap("PageDown",PageDown)
	beego.Run()
}

func Pageup(a int) int {
	if a<=1{
		return a
	}
	return a-1
}
func PageDown(page,pageindex int) int {
	if pageindex >=page{
		return page
	}
	return pageindex+1
}