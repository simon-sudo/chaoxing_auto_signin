package main

import (
	"fmt"
	"github.com/gocolly/colly"
	"go_dev/chaoxing_auto_signIn/config"
	"log"
	"regexp"
	"strings"
)

func main() {
	//登录页
	const loginUrl = "http://passport2.chaoxing.com/api/login?"
	loginInfoMap := map[string]string{
		"name":   config.Name,
		"pwd":    config.Pwd,
		"fid":    config.Fid,
		"verify": config.Verify,
	}
	//登录
	loginOp := CreateCollector()
	err := loginOp.Post(loginUrl, loginInfoMap)
	if err != nil {
		log.Println("登陆出错：", err)
		return
	}

	coursePage := loginOp.Clone()
	if err != nil {
		log.Println(err)
		return
	}
	//处理活动页面
	course := config.Course{
		CourseId: config.CourseId,
		ClassId:  config.ClassId,
	}
	coursePage.OnHTML("#startList .Mct", func(element *colly.HTMLElement) {
		s := element.Attr("onclick")
		reg1 := regexp.MustCompile(`\(\d+`)
		reg2 := regexp.MustCompile(`,\d+,`)
		activeId := reg1.FindString(s)
		activeType := reg2.FindString(s)
		activeId = activeId[1:len(activeId)]
		activeType = activeType[1 : len(activeType)-1]
		fmt.Println("处理活动",activeId,activeType)
		activeOp := coursePage.Clone()
		handleActive(activeOp, activeType, "https://mobilelearn.chaoxing.com/widget/sign/pcStuSignController/preSign?activeId="+activeId+"&classId="+course.ClassId+"&fid="+loginInfoMap["fid"]+"&courseId="+course.CourseId)
	})

	err = coursePage.Visit("https://mobilelearn.chaoxing.com/widget/pcpick/stu/index?courseId=" + course.CourseId + "&jclassId=" + course.ClassId)
}

func CreateCollector() *colly.Collector {
	coll := colly.NewCollector()
	coll.AllowURLRevisit = true
	coll.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept-Language", "zh,zh-TW;q=0.9,en-US;q=0.8,en;q=0.7,zh-CN;q=0.6,ru;q=0.5")
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.122 Safari/537.36")
		r.Headers.Set("Referer", "http://passport2.chaoxing.com/login?fid=&refer=http://i.mooc.chaoxing.com")
		r.Headers.Set("Proxy-Connection", "keep-alive")
	})
	return coll
}

func handleActive(activeOp *colly.Collector, activeType, url string) {
	if activeType == "2" {
		fmt.Println("签到ing")
		//签到活动
		resBody := string(normalSignin(activeOp, url).Body)
		if strings.Index(resBody,"签到成功")!=-1 {
			fmt.Println("签到成功！")
		} else if strings.Index(resBody,"拍照签到")!=-1{
			fmt.Println("开始拍照签到！")

		} else if strings.Index(resBody,"手势签到")!=-1{
			fmt.Println("开始手势签到")
		}
	}
}
func normalSignin(activeOp *colly.Collector, url string) *colly.Response {
	var res *colly.Response
	activeOp.OnResponse(func(resp *colly.Response) {
		res = resp
	})
	activeOp.Visit(url)
	return res
}
func photoSignin(activeOp *colly.Collector, url string) *colly.Response {
	var res *colly.Response
	activeOp.OnResponse(func(resp *colly.Response) {
		res = resp
	})
	activeOp.Visit(url)
	return res
}