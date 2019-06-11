package controllers

import (
	"GoodsShop/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/utils"
	"math"
	"regexp"
)

type UserController struct {
	beego.Controller
}
//显示登入页面
func(c*UserController) ShowLogin() {
	userName:=c.Ctx.GetCookie("userName")
	if userName!=""{
		c.Data["name"]=userName
		c.Data["checked"]="checked"
	}
	c.TplName="login.html"
}
//实现登入
func(c*UserController) HanderLogin() {
	//获取数据
	var user models.User
	user.Name = c.GetString("username")
	pwd:=c.GetString("pwd")
	remeberme:=c.GetString("checkbox")
	//校验

	if user.Name==""||pwd==""{
		c.Data["err"]="账号或密码不能为空"
		c.TplName="login.html"
		return
	}
	o:=orm.NewOrm()
	err:=o.Read(&user,"Name")
	if err!=nil{
		c.Data["err"]="用户名/密码 错误"
		c.TplName="login.html"
		return
	}
	if user.PassWord!=pwd{
		c.Data["err"]="用户名/密码 错误"
		c.TplName="login.html"
		return
	}
	if !user.Active{
		c.Ctx.WriteString("未激活")
		return
	}
	if remeberme=="on"{
		c.Ctx.SetCookie("userName",user.Name,60*60*24*7)
	}else {
		c.Ctx.SetCookie("userName",user.Name,-1)
	}
	//处理数据
	c.SetSession("userName",user.Name)
	//返回视图
	c.Redirect("/",302)
}
//显示注册页面
func(c*UserController) ShowRegister() {
	c.TplName="register.html"
}
//实现注册
func(c*UserController) HanderRegister() {
	//获取数据
	var user models.User
	user.Name=c.GetString("user_name")
	user.PassWord=c.GetString("pwd")
	user.Email=c.GetString("email")

	agree:=c.GetString("allow")

	//校验数据
	if agree!="on"{
		c.Data["err"]="数据不完整"
		c.TplName="register.html"
		return
	}


	if user.Name==""||user.PassWord==""||user.Email==""{
		c.Data["err"]="数据不完整"
		c.TplName="register.html"
		return
	}

	reg,_:=regexp.Compile(`[\w!#$%&"*+/=?^_"{|}~-]+(?:\.[\w!#$%&"*+/=?^_"{|}~-]+)*@(?:[\w](?:[\w-]*[\w])?\.)+[\w](?:[\w-]*[\w])?`)
	str:=reg.FindString(user.Email)
	if str==""{
		c.Data["err"]="邮箱格式不正确"
		c.TplName="register.html"
		return

	}
	//发送激活邮件
	err:=SendEmail(user.Email,user.Name)
	if err!=nil{
		c.Data["err"]="不知道为什么邮件发送失败"+err.Error()
		c.TplName="register.html"
		return
	}
	//处理数据

	o:=orm.NewOrm()
	_,err=o.Insert(&user)
	if err!=nil{
		c.Data["err"]="注册失败"
		c.TplName="register.html"
		return
	}

	c.Ctx.WriteString(`<h1>请前往邮箱激活<br><a href="/">主页</a></h1>`)

}
//激活账户
func(c*UserController) ShowActivate() {
	//获取数据
	userName:=c.GetString("name")
	var user models.User
	user.Name=userName
	o:=orm.NewOrm()
	err:=o.Read(&user,"Name")
	if err!=nil{
		c.Ctx.WriteString("用户不存在")
		return
	}
	user.Active=true
	_,err=o.Update(&user)
	if err!=nil{
		c.Ctx.WriteString("用户激活失败")
		return
	}
	c.Redirect("/login",302)
}
//发送激活邮件
func SendEmail(ToSomeBody string,name string)(err error){
	emailconfig:=`{"username":"youEmail","password":"youPassword","host":"host","port":25}`
	emailconn:=utils.NewEMail(emailconfig)
	beego.Info(emailconn)
	emailconn.From="youEmail"
	emailconn.To=[]string{ToSomeBody}
	emailconn.Subject="用户注册"
	emailconn.Text="激活邮件"+"http://127.0.0.1:8080/activate?name="+name
	err=emailconn.Send()
	return
}
//退出登入
func(c*UserController) ShowLogout() {
	c.DelSession("userName")
	c.Redirect("/login",302)
}
//显示用户中心
func(c*UserController) Showcenter() {
	userName:=GetNane(&c.Controller)
	var addrtemp models.Address
	orm.NewOrm().QueryTable("Address").RelatedSel("User").Filter("User__Name",userName).Filter("Isdefault",true).One(&addrtemp)

	GetCart(&c.Controller)
	c.Data["addr"]=addrtemp
	c.Data["center"]="active"
	c.Layout="user_layout.html"
	c.TplName="user_center_info.html"

}
//显示用户收货地址
func(c*UserController) ShowSize() {
	userName:=GetNane(&c.Controller)
	var addrtemp models.Address
	orm.NewOrm().QueryTable("Address").RelatedSel("User").Filter("User__Name",userName).Filter("Isdefault",true).One(&addrtemp)

	GetCart(&c.Controller)
	c.Data["addr"]=addrtemp
	c.Data["size"]="active"
	c.Layout="user_layout.html"
	c.TplName="user_center_site.html"


}
//添加收货地址
func(c*UserController) HanderSize() {
	//获取数据
	userName:=c.GetSession("userName").(string)
	var addr models.Address
	addr.Receiver=c.GetString("Receiver")
	addr.Addr=c.GetString("Addr")
	addr.Zipcode=c.GetString("Zipcode")
	addr.Phone=c.GetString("Phone")
	//校验数据
	if addr.Receiver==""||addr.Addr==""||addr.Zipcode==""||addr.Phone==""{
		c.Data["err"]="信息不完整"
		c.ShowSize()
		return
	}
	//存储数据
	o:=orm.NewOrm()
	var user models.User
	var addrtemp models.Address
	err:=o.QueryTable("Address").RelatedSel("User").Filter("User__Name",userName).Filter("Isdefault",true).One(&addrtemp)
	if err!=nil{
		addr.Isdefault=true
	}else {
		addr.Isdefault=false
	}
	user.Name=userName
	o.Read(&user,"Name")
	addr.User=&user
	_,err=o.Insert(&addr)
	if err!=nil{
		c.Data["err"]="地址添加失败"
		c.ShowSize()
		return
	}
	c.Redirect("/user/center",302)



}
//显示用户订单
func(c*UserController) ShowOrder() {
	GetNane(&c.Controller)
	goodsInfo:=make(map[int]interface{})
	var orderInfos  []models.OrderInfo
	o:=orm.NewOrm()
	//获取总页数
	count,_:=o.QueryTable("OrderInfo").Count()
	pageSize:=2
	page:=int(math.Ceil(float64(count)/float64(pageSize)))
	pageIndex,err:=c.GetInt("pageIndex")
	if err!=nil{
		pageIndex=1
	}
	if pageIndex<1{
		pageIndex=1
	}
	if pageIndex>=page{
		pageIndex=page
		pageSize=int(count)-(pageIndex-1)*pageSize
	}
	// 获取显示的5页
	pages:=PageTool(pageIndex,page)

	//订单排序优先显示最近订单
	o.QueryTable("OrderInfo").OrderBy("Id").Limit(pageSize,int(count)-pageIndex*pageSize).All(&orderInfos)
	for i, j := 0, len(orderInfos)-1; i < j; i, j = i+1, j-1 {
		orderInfos[i], orderInfos[j] = orderInfos[j], orderInfos[i]
	}
	for i,v:=range orderInfos{
		var orderGoods []models.OrderGoods

		temp:=make(map[string]interface{})
		o.QueryTable("OrderGoods").RelatedSel("GoodsSKU","OrderInfo").Filter("OrderInfo__Id",v.Id).All(&orderGoods)


		temp["ordergoods"]=orderGoods
		temp["orderinfo"]=v
		goodsInfo[i]=temp

	}
	c.Data["pageIndex"]=pageIndex
	c.Data["page"]=page
	c.Data["pages"]=pages
	c.Data["goodsinfo"]=goodsInfo
	c.Layout="user_layout.html"
	c.TplName="user_center_order.html"
}