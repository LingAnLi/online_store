package routers

import (
	"GoodsShop/controllers"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego"
)

func init() {
	//用户中心过滤非在线用户(游客？)
	beego.InsertFilter("/user/*",beego.BeforeExec,FnUser)
	//登入
	beego.Router("/login",&controllers.UserController{},"get:ShowLogin;post:HanderLogin")
	//注册
	beego.Router("/register",&controllers.UserController{},"get:ShowRegister;post:HanderRegister")
	//激活
	beego.Router("/activate",&controllers.UserController{},"get:ShowActivate")
	//退出
	beego.Router("/logout",&controllers.UserController{},"get:ShowLogout")
	//用户中心
	beego.Router("/user/center",&controllers.UserController{},"get:Showcenter")
	//用户中心 收货地址
	beego.Router("/user/centerSize",&controllers.UserController{},"get:ShowSize;post:HanderSize")
	//用户中心全部订单
	beego.Router("/user/centerOrder",&controllers.UserController{},"get:ShowOrder")

	// 主页
    beego.Router("/", &controllers.GoodsController{},"get:ShowIndex")
	//商品列表
	beego.Router("/list",&controllers.GoodsController{},"get:ShowList")
	//商品详细信息
	beego.Router("/detail",&controllers.GoodsController{},"get:ShowDetail")
	//搜索商品
	beego.Router("/goodsSearch",&controllers.GoodsController{},"get:ShowSearch")
	//添加购物车
	beego.Router("/cart",&controllers.CartController {},"post:AddCart")
	//显示购物车
	beego.Router("/user/cart",&controllers.CartController{},"get:ShowCart")
	//删除购物车商品
	beego.Router("/user/delGoods",&controllers.CartController{},"post:DelGoods")
	//结算界面
	beego.Router("/user/showOrder",&controllers.OrderController{},"post:HanderOrder")
	//提交结算
	beego.Router("/user/addOrder",&controllers.OrderController{},"post:HanderAddOrder")
	//付款
	beego.Router("/user/pay",&controllers.OrderController{},"get:ShowPay")
	//付款完成
	beego.Router("/user/payok",&controllers.OrderController{},"get:SetPay")

}
var FnUser = func(c*context.Context){
	userName:=c.Input.Session("userName")
	if userName==nil{
		c.Redirect(302,"/login")
		return
	}
}