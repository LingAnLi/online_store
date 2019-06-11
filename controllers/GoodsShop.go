package controllers

import (
	"GoodsShop/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"math"
)

type GoodsController struct {
	beego.Controller
}
//首页分类展示商品
type IndexGoodsType struct {
	MyType	string
	MyImg	string
	GoodsImgBanner        []models.IndexTypeGoodsBanner
	GoodsTxtBanner        []models.IndexTypeGoodsBanner
}
//获取登入用户名
func GetNane(c*beego.Controller)  string{
	userName:=c.GetSession("userName")
	if userName==nil{
		c.Data["name"]=""
		return ""
	}
		c.Data["name"]=userName.(string)
	return  userName.(string)
}
//首页显示
func (c *GoodsController) ShowIndex() {
	GetNane(&c.Controller)
	var goodsTypes []models.GoodsType
	var goodsbanners []models.IndexGoodsBanner
	var PromotionBanner []models.IndexPromotionBanner

	goods:= make(map[string]IndexGoodsType)

	o:=orm.NewOrm()
	o.QueryTable("GoodsType").All(&goodsTypes)
	o.QueryTable("IndexGoodsBanner").OrderBy("Index").All(&goodsbanners)
	o.QueryTable("IndexPromotionBanner").All(&PromotionBanner)
	for _,v:=range goodsTypes{
		//查询首页展示商品
		var imgbanner []models.IndexTypeGoodsBanner
		var txtbanner []models.IndexTypeGoodsBanner
		o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsType","GoodsSKU").Filter("GoodsType__Name",v.Name).Filter("DisplayType",1).All(&imgbanner)
		o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsType","GoodsSKU").Filter("GoodsType__Name",v.Name).Filter("DisplayType",0).All(&txtbanner)

		goods[v.Name]=IndexGoodsType{v.Name,v.Image,imgbanner,txtbanner}
	}

   	GetCart(&c.Controller)
	//返回视图
	c.Data["goods"]=goods
	c.Data["PromotionBanner"]=PromotionBanner
	c.Data["goodsbanners"]=goodsbanners
	c.Data["goodsTypes"]=goodsTypes
	c.Layout="shop_layout.html"
	c.TplName = "index.html"
}
//显示商品列表
func(c*GoodsController) ShowList() {
	GetNane(&c.Controller)
	//获取当前页码
	pageIndex,err:=c.GetInt("pageIndex")
	if err!=nil{
		pageIndex=1
	}

	//获取分类
	goodsTypeName:=c.GetString("name")
	var goodsSku []models.GoodsSKU
	var goodsTypes []models.GoodsType
	o:=orm.NewOrm()
	// 查找所有分类
	o.QueryTable("GoodsType").All(&goodsTypes)
	//详情
	var count int64//总记录
	var pageSize int =4
	var page int
	//商品分类排序
	sort:=c.GetString("sort")
	if sort==""{
		if goodsTypeName==""{
			count,_=o.QueryTable("GoodsSKU").Count()
			page=int( math.Ceil(float64(count)/float64(pageSize) ) )
			o.QueryTable("GoodsSKU").Limit(pageSize,(pageIndex-1)*pageSize).RelatedSel("GoodsType").All(&goodsSku)
		}else {
			count,_=o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Name",goodsTypeName).Count()
			page=int( math.Ceil(float64(count)/float64(pageSize) ) )
			o.QueryTable("GoodsSKU").Limit(pageSize,(pageIndex-1)*pageSize).RelatedSel("GoodsType").Filter("GoodsType__Name",goodsTypeName).All(&goodsSku)
		}
	} else if sort=="price"{
		if goodsTypeName==""{
			count,_=o.QueryTable("GoodsSKU").Count()
			page=int( math.Ceil(float64(count)/float64(pageSize) ) )
			o.QueryTable("GoodsSKU").Limit(pageSize,(pageIndex-1)*pageSize).RelatedSel("GoodsType").OrderBy("Price").All(&goodsSku)
		}else {
			count,_=o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Name",goodsTypeName).Count()
			page=int( math.Ceil(float64(count)/float64(pageSize) ) )
			o.QueryTable("GoodsSKU").Limit(pageSize,(pageIndex-1)*pageSize).RelatedSel("GoodsType").Filter("GoodsType__Name",goodsTypeName).OrderBy("Price").All(&goodsSku)
		}
	}else if sort=="sale"{
		if goodsTypeName==""{
			count,_=o.QueryTable("GoodsSKU").Count()
			page=int( math.Ceil(float64(count)/float64(pageSize) ) )
			o.QueryTable("GoodsSKU").Limit(pageSize,(pageIndex-1)*pageSize).RelatedSel("GoodsType").OrderBy("Sales").All(&goodsSku)
		}else {
			count,_=o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Name",goodsTypeName).Count()
			page=int( math.Ceil(float64(count)/float64(pageSize) ) )
			o.QueryTable("GoodsSKU").Limit(pageSize,(pageIndex-1)*pageSize).RelatedSel("GoodsType").Filter("GoodsType__Name",goodsTypeName).OrderBy("Sales").All(&goodsSku)
		}
	}

	//总页数
	//beego.Info(page)
	//新品推荐
	NewProduct(&c.Controller,goodsTypeName)
	//页码显示
	showPape:=PageTool(pageIndex,page)
	//购物车商品数量
	GetCart(&c.Controller)
	//返回视图
	c.Data["sort"]=sort
	c.Data["showPage"]=showPape
	c.Data["pageindex"]=pageIndex
	c.Data["page"]=page
	c.Data["typeName"]=goodsTypeName
	c.Data["sku"]=goodsSku
	c.Data["goodsTypes"]=goodsTypes
	c.Layout="shop_layout.html"
	c.TplName="list.html"
}
//显示商品详细信息
func(c*GoodsController) ShowDetail() {
	GetNane(&c.Controller)
	//获取数据
	typeName:=c.GetString("typeName")
	id,err:=c.GetInt("id")

	if err!=nil{
		beego.Info("商品不存在")
	}
	var sku models.GoodsSKU
	var goodsTypes []models.GoodsType
	sku.Id=id
	o:=orm.NewOrm()
	err=o.QueryTable("GoodsSKU").RelatedSel("Goods").Filter("Id",id).One(&sku)
	if err!=nil{
		//beego.Info("商品不存在")
		c.Redirect("/",302)
		return
	}
	//获取所分类
	o.QueryTable("GoodsType").All(&goodsTypes)
	//获取当前分类信息
	var myType models.GoodsType

	myType.Name=typeName
	err=o.Read(&myType,"Name")
	if err!=nil{
		beego.Info("读取分类错误",err)
		typeName=""
	}
	//新品推荐
	NewProduct(&c.Controller,typeName)
	//购物车数量
	GetCart(&c.Controller)
	//返回视图
	c.Data["typeName"]=typeName
	c.Data["goodsTypes"]=goodsTypes
	c.Data["sku"]=sku
	c.Layout="shop_layout.html"
	c.TplName="detail.html"
}
//新品推荐
func NewProduct(c*beego.Controller,typeName string){
	//beego.Info(typeName)
	o:=orm.NewOrm()
	var sku []models.GoodsSKU
	//获取当前分类的最新两条数据
	if typeName!=""{
		o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Name",typeName).Limit(2,0).OrderBy("Time").All(&sku)
	}else{
		o.QueryTable("GoodsSKU").Limit(2,0).OrderBy("Time").All(&sku)
	}
	c.Data["NewProduct"]=sku
}
//列表页码显示
func PageTool(pageIndex,page int) []int {
	//总页数不足5
	var showPage []int
	if page<=5{
		showPage=make([]int,page)
		for i,_:=range showPage{
			showPage[i]=i+1
		}
	}else if page-pageIndex<3{
		showPage=[]int{page-4,page-3,page-2,page-1,page}//总页数-当前页<3
	}else if pageIndex<=3{
		showPage=[]int{1,2,3,4,5}
	}else{
		showPage=[]int{pageIndex-2,pageIndex-1,pageIndex,pageIndex+1,pageIndex+2}
	}
	return showPage

}
//显示搜索结果
func(c*GoodsController) ShowSearch() {
	GetNane(&c.Controller)
	//获取当前页码
	pageIndex,err:=c.GetInt("pageIndex")
	if err!=nil{
		pageIndex=1
	}
	//获取分类
	var goodsSku []models.GoodsSKU
	var goodsTypes []models.GoodsType
	o:=orm.NewOrm()
	// 查找所有分类
	o.QueryTable("GoodsType").All(&goodsTypes)
	//获取搜索数据
	goodsName:=c.GetString("goodsName")
	//详情
	var count int64//总记录
	var pageSize int =15
	var page int
	//商品分类排序
	sort:=c.GetString("sort")
	if sort==""{
			count,_=o.QueryTable("GoodsSKU").Filter("Name__icontains",goodsName).Count()
			page=int( math.Ceil(float64(count)/float64(pageSize) ) )
			o.QueryTable("GoodsSKU").Limit(pageSize,(pageIndex-1)*pageSize).RelatedSel("GoodsType").Filter("Name__icontains",goodsName).All(&goodsSku)
	} else if sort=="price"{
			count,_=o.QueryTable("GoodsSKU").Filter("Name__icontains",goodsName).Count()
			page=int( math.Ceil(float64(count)/float64(pageSize) ) )
			o.QueryTable("GoodsSKU").Limit(pageSize,(pageIndex-1)*pageSize).RelatedSel("GoodsType").OrderBy("Price").Filter("Name__icontains",goodsName).All(&goodsSku)
	}else if sort=="sale"{
			count,_=o.QueryTable("GoodsSKU").Filter("Name__icontains",goodsName).Count()
			page=int( math.Ceil(float64(count)/float64(pageSize) ) )
			o.QueryTable("GoodsSKU").Limit(pageSize,(pageIndex-1)*pageSize).RelatedSel("GoodsType").OrderBy("Sales").Filter("Name__icontains",goodsName).All(&goodsSku)
			}

	//总页数
	beego.Info(page)
	//新品推荐
	NewProduct(&c.Controller,"")
	//页码显示
	showPape:=PageTool(pageIndex,page)

	GetCart(&c.Controller)
	c.Data["goodsName"]=goodsName
	c.Data["sort"]=sort
	c.Data["showPage"]=showPape
	c.Data["pageindex"]=pageIndex
	c.Data["page"]=page
	c.Data["sku"]=goodsSku
	c.Data["goodsTypes"]=goodsTypes
	c.Layout="shop_layout.html"
	c.TplName="goods_search.html"
}