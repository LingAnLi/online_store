package controllers

import (
	"GoodsShop/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/garyburd/redigo/redis"
	"strconv"
)

type CartController struct {
	beego.Controller
}
//添加购物车商品id 数量
func(c*CartController) AddCart() {
	//获取数据
	skuid,err1:=c.GetInt("skuid")
	count,err2:=c.GetInt("count")
	resp:=make(map[string]interface{})
	defer c.ServeJSON()
	if err1!=nil||err2!=nil{
		beego.Info("获取数据错误")
	}
	beego.Info(skuid,count)
	userName:=c.GetSession("userName")
	if userName==nil{
		resp["code"]=4
		resp["msg"]="用户未登录"
		c.Data["json"]=resp
		return
	}
	var user models.User
	user.Name=userName.(string)
	orm.NewOrm().Read(&user,"Name")
	//处理数据
	conn,err:=redis.Dial("tcp","127.0.0.1:6379")
	defer conn.Close()
	if err!=nil{
		beego.Info(err)
		resp["code"]="4"
		resp["msg"]="连接数据库错误"
		c.Data["json"]=resp
	}
	//存入redis
	preCount,err:=redis.Int(conn.Do("hget","cart_"+strconv.Itoa(user.Id),skuid))
	_,err1=conn.Do("hset","cart_"+strconv.Itoa(user.Id),skuid,count+preCount)
	res,err2:=redis.Int(conn.Do("hlen","cart_"+strconv.Itoa(user.Id)))

	//返回数据
	if err1==nil&&err2==nil{
		resp["code"]="5"
		resp["msg"]="ok"
		resp["data"]=res
		c.Data["json"]=resp
		return
	}
	resp["code"]="2"
	resp["msg"]="加入购物车失败"
	c.Data["json"]=resp


}
//删除购物车商品
func(c*CartController) DelGoods() {
	//获取商品ID
	id,err:=c.GetInt("skuid")
	resp:=make(map[string]interface{})
	defer  c.ServeJSON()
	if err!=nil{
		resp["code"]=1
		resp["errmsg"]="Get SkuId ERROR"
		c.Data["json"]=resp
		return
	}
	conn,err:=redis.Dial("tcp","127.0.0.1:6379")
	if err!=nil{
		resp["code"]=2
		resp["errmsg"]="Dial Redis ERROR"
		c.Data["json"]=resp

		return
	}
	userName:=c.GetSession("userName")
	if userName==nil{
		//未登入————>已经过滤
		resp["code"]=3
		resp["errmsg"]="___"
		c.Data["json"]=resp

		return
	}
	var user models.User
	user.Name=userName.(string)
	orm.NewOrm().Read(&user,"Name")
	//删除
	_,err=conn.Do("hdel","cart_"+strconv.Itoa(user.Id),id)
	if err !=nil{
		//删除失败
		beego.Error("Del ERROR",err)
		resp["code"]=4
		resp["errmsg"]="Del ERROR"
		c.Data["json"]=resp

		return
	}
	resp["code"]=5
	resp["errmsg"]="ok"
	c.Data["json"]=resp
}
//显示购物车
func(c*CartController) ShowCart() {
	userName:=GetNane(&c.Controller)
	var user models.User
	user.Name=userName
	o:=orm.NewOrm()
	o.Read(&user,"Name")
	//从redis中获取购物车数据
	conn,err:=redis.Dial("tcp","127.0.0.1:6379")
	if err!=nil{
		beego.Error("连接失败",err)
		return
	}
	defer conn.Close()
	goodsSku,_:=redis.IntMap(conn.Do("hgetall","cart_"+strconv.Itoa(user.Id)))
	allGoods:=make([]interface{},0)
	//拿到购物车商品的数据
	for i,v:=range goodsSku{
		skuid,_:=strconv.Atoi(i)
		temp:= map[string]interface{}{}
		var goodsSku models.GoodsSKU
		goodsSku.Id=skuid
		o.Read(&goodsSku)
		temp["goods"]=goodsSku
		temp["count"]=v
		allGoods=append(allGoods,temp)
	}
	//返回数据
	c.Data["goods"]=allGoods
	c.Layout="cart_layout.html"
	c.TplName="cart.html"
}
// 显示购物车商品数量
func GetCart(c*beego.Controller) int {
	userName:=c.GetSession("userName")
	if userName==nil {
		c.Data["cart"]=0
		return 0
	}
	var user models.User
	user.Name=userName.(string)
	orm.NewOrm().Read(&user,"Name")
	conn,_:=redis.Dial("tcp","127.0.0.1:6379")
	defer conn.Close()
	res,_:=redis.Int(conn.Do("hlen","cart_"+strconv.Itoa(user.Id)))

	c.Data["cart"]=res
	return res
}