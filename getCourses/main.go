package main

import (
	"github.com/gocolly/colly"
	"go_dev/chaoxing_auto_signIn/config"
	"log"
	"os"
)

func main() {
	//登录页
	const loginUrl = "http://passport2.chaoxing.com/api/login?"
	loginInfoMap := map[string]string{
		"name":		config.Name,
		"pwd": 		config.Pwd,
		"verify":	config.Verify,
	}
	//登录
	loginOp := CreateCollector()
	err := loginOp.Post(loginUrl, loginInfoMap)
	if err!=nil{
		log.Println("登陆出错：" , err)
		return
	}
	courses := []config.Course{}
	coursePage := loginOp.Clone()
	if err!=nil{
		log.Println(err)
		return
	}
	coursePage.OnHTML("div.ulDiv ul li", func(element *colly.HTMLElement) {
		s := element.ChildAttrs("input","value")
		courseName := element.ChildAttr("h3 a","title")
		if len(s)!=0{
			courses = append(courses, config.Course{s[0],s[1],courseName})
		}
	})
	file ,err:= os.Create("a.txt")
	if err!=nil{
		log.Println(err)
	}
	err = coursePage.Visit("http://mooc1-1.chaoxing.com/visit/interaction")
	for _,v := range courses{
		file.WriteString("courseId="+v.CourseId+"\n")
		file.WriteString("classId="+v.ClassId+"\n")
		file.WriteString("courseName="+v.CourseName+"\n\n")
	}
	file.Close()
}


func CreateCollector() *colly.Collector {
	coll := colly.NewCollector()
	coll.AllowURLRevisit = true
	coll.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept-Language", "zh,zh-TW;q=0.9,en-US;q=0.8,en;q=0.7,zh-CN;q=0.6,ru;q=0.5")
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.122 Safari/537.36")
		r.Headers.Set("Referer", "http://passport2.chaoxing.com/login?fid=&refer=http://i.mooc.chaoxing.com")
		r.Headers.Set("Proxy-Connection" , "keep-alive")
	})
	return coll
}