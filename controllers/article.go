package controllers

import (
	"github.com/astaxie/beego"
	"path"
	"time"
	"github.com/astaxie/beego/orm"
	"newsWeb/models"
	"math"
	"strconv"
)

type ArticleController struct {
	beego.Controller
}

//展示文章列表页
func (this *ArticleController) ShowArticleList() {

	userName := this.GetSession("userName")
	if userName == nil {
		this.Redirect("/login", 302)
	}
	this.Data["userName"] = userName.(string)
	//查询数据库，拿出数据，传递给视图
	//获取orm对象
	o := orm.NewOrm()
	//获取查询对象
	var articles []models.Article
	//查询
	//queryseter  高级查询使用的数据类型
	qs := o.QueryTable("Article")
	//查询所有的文章
	//qs.All(&articles)//select * from article

	//实现分页
	//获取总记录数和总页数
	typeName := this.GetString("select")
	var count int64
	if typeName == "" {
		count, _ = qs.RelatedSel("ArticleType").Count()
	} else {
		count, _ = qs.RelatedSel("ArticleType").Filter("ArticleType__TypeName", typeName).Count()
	}

	pageSize := int64(2)

	pageCount := float64(count) / float64(pageSize)

	pageCount = math.Ceil(pageCount)

	//向上取整
	//把数据传递给视图
	this.Data["count"] = count
	this.Data["pageCount"] = pageCount

	//获取首页末页数据
	pageIndex, err := this.GetInt("pageIndex")
	if err != nil {
		pageIndex = 1
	}
	//获取分页的数据
	start := pageSize * (int64(pageIndex) - 1)

	//根据传递的类型获取相应的文章
	//获取数据
	this.Data["typeName"] = typeName
	if typeName == "" {
		qs.Limit(pageSize, start).RelatedSel("ArticleType").All(&articles) //queryseter
	} else {
		qs.Limit(pageSize, start).RelatedSel("ArticleType").Filter("ArticleType__TypeName", typeName).All(&articles)
	}


	var articleTypes []models.ArticleType
	o.QueryTable("ArticleType").All(&articleTypes)
	this.Data["articleTypes"] = articleTypes
	this.Data["pageIndex"] = pageIndex
	this.Data["articles"] = articles

	this.Layout = "layout.html"
	this.TplName = "index.html"
}

//展示添加文章页面
func (this *ArticleController) ShowAddArticle() {
	//获取所有类型 传递给试图
	o := orm.NewOrm()
	var articleTpes []models.ArticleType

	o.QueryTable("ArticleType").All(&articleTpes)
	this.Data["articleTypes"] = articleTpes

	this.Layout = "layout.html"
	this.TplName = "add.html"
}

//处理添加文章业务
func (this *ArticleController) HandleAddArticle() {
	//接受数据
	artileName := this.GetString("articleName")
	content := this.GetString("content")
	//校验数据
	if artileName == "" || content == "" {
		this.Data["errmsg"] = "文章标题或内容不能为空"
		this.TplName = "add.html"
		return
	}

	typeName := this.GetString("select")

	//接收图片
	file, head, err := this.GetFile("uploadname")
	if err != nil {
		this.Data["errmsg"] = "获取文件失败"
		this.TplName = "add.html"
		return
	}
	defer file.Close()
	//1.判断文件大小
	if head.Size > 500000 {
		this.Data["errmsg"] = "文件太大，上传失败"
		this.TplName = "add.html"
		return
	}

	//2.判断图片格式
	//1.jpg
	fileExt := path.Ext(head.Filename)
	if fileExt != ".jpg" && fileExt != ".png" && fileExt != ".jpeg" {
		this.Data["errmsg"] = "文件格式不正确，请重新上传"
		this.TplName = "add.html"
		return
	}

	//3.文件名防止重复
	fileName := time.Now().Format("2006-01-02-15-04-05") + fileExt
	this.SaveToFile("uploadname", "./static/image/"+fileName)

	//处理数据
	//数据库的插入操作
	//获取orm对象
	o := orm.NewOrm()
	//获取插入对象
	var article models.Article
	//给插入对象赋值
	article.Title = artileName
	article.Content = content
	article.Image = "/static/image/" + fileName

	var articleType models.ArticleType
	articleType.TypeName = typeName
	o.Read(&articleType, "TypeName")
	article.ArticleType = &articleType

	//插入
	_, err = o.Insert(&article)
	if err != nil {
		this.Data["errmsg"] = "添加文章失败，请重新添加"
		this.TplName = "add.html"
		return
	}
	//返回页面
	this.Redirect("/article/articleList", 302)
}

//展示文章详情页
func (this *ArticleController) ShowArticleDetail() {
	//获取数据
	articleId, err := this.GetInt("id")
	//校验数据
	if err != nil {
		this.Data["errmsg"] = "请求路径错误"
		this.TplName = "index.html"
		return
	}

	//获取到文章

	//处理数据
	//查询数据
	//获取orm对象
	o := orm.NewOrm()
	//获取查询对象
	var article models.Article
	//给查询条件赋值
	article.Id = articleId
	//查询
	err = o.Read(&article)
	if err != nil {
		this.Data["errmsg"] = "请求路径错误"
		this.TplName = "index.html"
		return
	}

	m2m := o.QueryM2M(&article, "Users")
	//获取要插入对象
	var user models.User
	userName := this.GetSession("userName")
	user.UserName = userName.(string)
	o.Read(&user, "UserName")
	m2m.Add(user)

	//第1种多对多查询
	o.LoadRelated(&article, "Users")

	//第二种多对多查询
	var users []models.User
	o.QueryTable("User").Filter("Articles__Article__Id", articleId).Distinct().All(&users)
	this.Data["Users"] = users
	//返回数据
	this.Data["article"] = article
	this.TplName = "content.html"
}

func (this *ArticleController) ShowUpdateArticle() {

	articleId, err := this.GetInt("id")
	errmsg := this.GetString("errmsg")
	if errmsg != "" {
		this.Data["errmsg"] = errmsg
	}
	if err != nil {
		beego.Error("请求路径错误")
		this.Redirect("/article/articleList?errmsg", 302)
		return
	}
	o := orm.NewOrm()

	var article models.Article
	article.Id = articleId
	o.Read(&article)

	this.Data["article"] = article
	this.TplName = "update.html"

}

//文件上传函数
func UploadFile(this *ArticleController, filePath string) string {
	//接收图片
	file, head, err := this.GetFile(filePath)
	if err != nil {
		this.Data["errmsg"] = "获取文件失败"
		this.TplName = "add.html"
		return ""
	}
	defer file.Close()
	//1.判断文件大小
	if head.Size > 500000 {
		this.Data["errmsg"] = "文件太大，上传失败"
		this.TplName = "add.html"
		return ""
	}

	//2.判断图片格式
	//1.jpg
	fileExt := path.Ext(head.Filename)
	if fileExt != ".jpg" && fileExt != ".png" && fileExt != ".jpeg" {
		this.Data["errmsg"] = "文件格式不正确，请重新上传"
		this.TplName = "add.html"
		return ""
	}

	//3.文件名防止重复
	fileName := time.Now().Format("2006-01-02-15-04-05") + fileExt
	this.SaveToFile(filePath, "./static/image/"+fileName)
	return "/static/image/" + fileName
}

func (this *ArticleController) HandleUpdateArticle() {
	articleName := this.GetString("articleName")
	content := this.GetString("content")
	fileName := UploadFile(this, "uploadname")

	articleId, err2 := this.GetInt("id")
	if articleName == "" || content == "" || fileName == "" || err2 != nil {
		errmsg := "内容不能为空"
		this.Redirect("/article/updateArticle?id="+strconv.Itoa(articleId)+"&errmsg"+errmsg, 302)
		return
	}
	o := orm.NewOrm()
	var article models.Article
	article.Id = articleId
	err := o.Read(&article)
	if err != nil {
		errmsg := "更新文章不存在"
		this.Redirect("/article/updateArticle?id="+strconv.Itoa(articleId)+"&errmsg="+errmsg, 302)
		return
	}
	//给更新字段赋新值
	article.Title = articleName
	article.Content = content
	article.Image = fileName
	//更新
	o.Update(&article)

	//返回数据
	this.Redirect("/article/articleList", 302)

}

//删除业务处理

func (this *ArticleController) DeleteArticle() {
	articleId, err := this.GetInt("id")
	if err != nil {
		beego.Error("路径错误")
		this.Redirect("/article/articleList", 302)
		return
	}
	o := orm.NewOrm()
	var article models.Article

	article.Id = articleId
	_, err = o.Delete(&article)
	if err != nil {
		beego.Error("删除失败")
		this.Redirect("/article/articleList", 302)
		return

	}
	this.Redirect("/article/articleList", 302)
}

//展示添加类型

func (this *ArticleController) ShowAddType() {

	o := orm.NewOrm()
	var articleTypes []models.ArticleType
	qs := o.QueryTable("ArticleType")
	qs.All(&articleTypes)

	this.Data["articleTypes"] = articleTypes
	this.Layout = "layout.html"
	this.TplName = "addType.html"
}

func (this *ArticleController) HandleAddType() {
	typeName := this.GetString("typeName")
	if typeName == "" {
		this.Data["errmsg"] = "类型名不能为空"
		this.Redirect("/article/addType?errmsg=", 302)
		return
	}
	o := orm.NewOrm()
	var articleType models.ArticleType
	articleType.TypeName = typeName
	_, err := o.Insert(&articleType)
	if err != nil {
		this.Data["errmsg"] = "文章类型添加失败"
		this.TplName = "addType.html"
		return
	}

	this.Redirect("/article/addType", 302)

}

//删除类型
func (this *ArticleController) DeleteType() {
	typeId, err := this.GetInt("id")
	if err != nil {
		this.Data["errmsg"] = "删除失败"

		this.Redirect("/article/addType?errmsg=", 302)
		return
	}
	o := orm.NewOrm()
	var articleType models.ArticleType
	articleType.Id = typeId
	_, err = o.Delete(&articleType)
	if err != nil {
		this.Data["errmsg"] = "删除失败"
		this.Redirect("/article/addType?errmsg=", 302)
		return
	}
	this.Redirect("/article/addType", 302)
}

//下拉框数据处理
func (this *ArticleController) HandleSelect() {
	typeName := this.GetString("select")

	if typeName == "" {
		beego.Error("下拉框获取数据失败")
		this.Redirect("/article/articleList", 302)
		return
	}

}
