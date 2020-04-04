package main

import (
	"encoding/json"
	"github.com/gocolly/colly"
	"io/ioutil"
	"log"
	"os"
	"strings"
)
type Config struct {
	Account string		`json:"Account"`
	Pwd string			`json:"Pwd"`
	Verify string		`json:"Verify"`
	Fid string			`json:"Fid"`
	Name string			`json:"Name"`
}
type Courses struct {
	CourseId []string 	`json:"CourseId"`
	ClassId []string	`json:"ClassId"`
	CourseName []string	`json:"CourseName"`
	Config
}
var conf Config
var courses Courses
func init() {
	configPath,_ := os.Getwd()
	configPath = configPath[:strings.LastIndex(configPath,"\\")]
	configPath += "/config/user.json"
	data,err := ioutil.ReadFile(configPath)
	if err!=nil{
		log.Println(err)
	}
	err = json.Unmarshal(data,&conf)
	courses.Config = conf
	if err!=nil{
		log.Println(err)
	}
}
func main() {
	//登录页
	const loginUrl = "http://passport2.chaoxing.com/api/login?"
	loginInfoMap := map[string]string{
		"name":		conf.Account,
		"pwd": 		conf.Pwd,
		"verify":	conf.Verify,
	}
	//登录
	loginOp := CreateCollector()
	err := loginOp.Post(loginUrl, loginInfoMap)
	if err!=nil{
		log.Println("登陆出错：" , err)
		return
	}
	coursePage := loginOp.Clone()
	if err!=nil{
		log.Println(err)
		return
	}
	coursePage.OnHTML("div.ulDiv ul li", func(element *colly.HTMLElement) {
		s := element.ChildAttrs("input","value")
		courseName := element.ChildAttr("h3 a","title")
		if len(s)!=0{
			courses.CourseId = append(courses.CourseId, s[0])
			courses.ClassId = append(courses.ClassId, s[1])
			courses.CourseName = append(courses.CourseName, courseName)
		}
	})
	path,_ := os.Getwd()
	path = path[:strings.LastIndex(path,"\\")]
	path += "/config/user.json"
	file ,err:= os.Create(path)
	if err!=nil{
		log.Println(err)
	}
	err = coursePage.Visit("http://mooc1-1.chaoxing.com/visit/interaction")
	coursesStr,err:= json.MarshalIndent(courses,"","   ")
	if err!=nil{
		log.Println(err)
	}
	file.WriteString(string(coursesStr))
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