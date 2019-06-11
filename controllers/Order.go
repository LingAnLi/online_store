package controllers

import (
	"GoodsShop/models"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/garyburd/redigo/redis"
	"github.com/smartwalle/alipay"
	"strconv"
	"strings"
	"time"
)

type OrderController struct {
	beego.Controller
}
//显示结算界面
func(c*OrderController) HanderOrder() {
	//获取用户信息
	userName:=GetNane(&c.Controller)
	var user models.User
	o:=orm.NewOrm()
	user.Name=userName
	o.Read(&user,"Name")
	//获取地址信息
	var addrs []models.Address
	o.QueryTable("Address").RelatedSel("User").Filter("User__Id",user.Id).All(&addrs)
	//beego.Info(addrs)
	//获取订单信息

	conn,err:=redis.Dial("tcp","127.0.0.1:6379")
	if err!=nil{
		beego.Info("redis Err",err)
		return
	}
	defer conn.Close()
	mycount,err:=c.GetInt("count")

	goods:=make(map[int]interface{})
	// 购物车或直接购买
	if err!=nil{
		// 购物车
		skuids:=c.GetStrings("skuid")
		if len(skuids)==0{
			c.Redirect("/user/cart",302)
			return
		}


		var goodsSku models.GoodsSKU
		//获取添加购物车商品的数据
		for i,v:=range skuids{
			id,_:=strconv.Atoi(v)
			goodsSku.Id= id
			o.Read(&goodsSku)
			count,err:=redis.Int(conn.Do("hget","cart_"+strconv.Itoa(user.Id),id))
			if err!=nil{
				c.Redirect("/user/cart",302)
				return
			}
			tmp:=make(map[string]interface{})
			tmp["count"]=count
			beego.Info(count)
			tmp["sku"]=goodsSku
			goods[i]=tmp
		}
		c.Data["skuids"]=skuids
	}else {
		var goodsSku models.GoodsSKU
		skuid,_:=c.GetInt("skuid")
		//添加直接购买商品的数据到购物车——————》偷懒啦
		redis.Int(conn.Do("hset","cart_"+strconv.Itoa(user.Id),skuid,mycount))

		goodsSku.Id=skuid
		o.Read(&goodsSku)
		tmp:=make(map[string]interface{})
		tmp["count"]=mycount
		tmp["sku"]=goodsSku
		goods[1]=tmp
		c.Data["skuids"]=[]int{skuid}
	}


	c.Data["goods"]=goods
	c.Data["addrs"]=addrs
	c.Layout="cart_layout.html"
	c.TplName="place_order.html"
}
//提交结算数据
func(c*OrderController) HanderAddOrder() {
	userName:=GetNane(&c.Controller)
	addrid,err1 :=c.GetInt("addrid")
	payid,err2	:=c.GetInt("payid")
	skuid:=c.GetString("skuids")
	json:=make(map[string]interface{})
	defer c.ServeJSON()


	if err1!=nil||err2!=nil{
		json["code"]=2
		json["errmsg"]="没有获取到支付或地址信息"
		c.Data["json"]=json
	}
	beego.Info(addrid,payid)
	skuids:=strings.Split(skuid[1:len(skuid)-1]," ")
	//beego.Info(skuids)
	if len(skuids)==0 {
		json["code"]=1
		json["errmsg"]="没有获取到skuid"
		c.Data["json"]=json
	}

	//处理数据
	o:=orm.NewOrm()
	var user models.User
	user.Name=userName
	var addr models.Address
	addr.Id=addrid
	o.Read(&addr)
	o.Read(&user,"Name")

	conn,err:=redis.Dial("tcp","127.0.0.1:6379")
	if err!=nil{
		beego.Info("连接redis失败",err)
		return
	}
	defer  conn.Close()

	TotalCount:=0
	TotalPrice:=0

	ordergoodss:=make([]models.OrderGoods,len(skuids))
	o.Begin()
	//获取商品数量/价格
	Stock:=make([]int,0)
	for i,v:=range skuids{

		skuid,_:=strconv.Atoi(v)
		count,err:=redis.Int(conn.Do("hget","cart_"+strconv.Itoa(user.Id),skuid))
		if err!=nil{
			json["code"]=4
			json["errmsg"]="读取商品数量错误"
			c.Data["json"]=json
			beego.Info("读取商品数量错误",err)
			return
		}

		var sku models.GoodsSKU
		sku.Id=skuid
		o.Read(&sku)

		if sku.Stock<count{
			json["code"]=3
			json["errmsg"]=sku.Name+"库存不足"
			c.Data["json"]=json
			return
		}

		if err!=nil {
			json["code"]=3
			json["errmsg"]="添加商品失败"
			c.Data["json"]=json
			o.Rollback()
			return
		}
		Stock=append(Stock,sku.Stock)
		x:=sku.Price*count
		TotalPrice+=x
		TotalCount+=count
		ordergoodss[i].GoodsSKU=&sku
		ordergoodss[i].Price=sku.Price
		ordergoodss[i].Count=count
		beego.Info(ordergoodss[i])

	}
	//订单表
	var orderInfo models.OrderInfo
	orderInfo.OrderId=time.Now().Format("20060102150405")+strconv.Itoa(user.Id)+strconv.Itoa(addr.Id)
	orderInfo.User=&user
	orderInfo.Address=&addr
	orderInfo.PayMethod=payid
	orderInfo.TotalCount=TotalCount//商品数量
	orderInfo.TransitPrice=10//运费
	orderInfo.TotalPrice=TotalPrice+10//价格


	_,err=o.Insert(&orderInfo)
	if err!=nil{
		o.Rollback()
		json["code"]=5
		json["errmsg"]="订单生成失败"
		c.Data["json"]=json
		return

	}
	//订单商品
	for _,ordergoods:=range ordergoodss{
		ordergoods.OrderInfo=&orderInfo

		_,err=o.Insert(&ordergoods)

		if err!=nil{

			o.Rollback()
			json["code"]=5
			json["errmsg"]="订单失败"
			c.Data["json"]=json
			return

		}
	}
	//并发导致出售商品>库存-----》》》多次校验库存

	o.Commit()
	//删除购物车
	for _,v:=range skuids{
		_,err=conn.Do("hdel","cart_"+strconv.Itoa(user.Id),v)
	}
	json["code"]=0
	json["errmsg"]="ok"
	c.Data["json"]=json

}
//支付
func(c*OrderController) ShowPay() {
	infoId,_:=c.GetInt("id")
	var orderInfo models.OrderInfo
	orderInfo.Id=infoId
	o:=orm.NewOrm()
	o.Read(&orderInfo)
    if orderInfo.PayMethod==3{
		url:=AilPay(orderInfo.OrderId,strconv.Itoa(orderInfo.TotalPrice)+".00",strconv.Itoa(infoId))
		c.Redirect(url,302)
	}
	c.Ctx.WriteString("暂不支持此方式")
}
//支付结果返回
func(c*OrderController) SetPay() {

	orderId:=c.GetString("out_trade_no")
	TradeNo:=c.GetString("trade_no")

	o:=orm.NewOrm()
	o.QueryTable("OrderInfo").Filter("OrderId",orderId).Update(orm.Params{"Orderstatus":1,"TradeNo":TradeNo})
	c.Redirect("/user/centerOrder",302)
}
//ailpay支付
func AilPay(OutTradeNo,TotalAmount,id string )string{
	var aliPublicKey = "公钥"
	var privateKey = "私钥"
	client:= alipay.New("appid", aliPublicKey, privateKey, false)
	var p = alipay.AliPayTradePagePay{}
	//p.NotifyURL = "http://xxx"
	p.ReturnURL = "http://127.0.0.1:8080/user/payok"
	p.Subject = "xxx购物"
	p.OutTradeNo = OutTradeNo
	p.TotalAmount = TotalAmount
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"

	var url, err = client.TradePagePay(p)
	if err != nil {
		fmt.Println(err)
	}
	beego.Info(url)
	 payURL := url.String()
	return payURL

}
