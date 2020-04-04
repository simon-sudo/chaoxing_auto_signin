package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"go_dev/chaoxing_auto_signIn/config"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	url2 "net/url"
	"os"
	"regexp"
	"strings"
)
var loginInfoMap = make(map[string]string)

func main() {
	//登录页
	const loginUrl = "http://passport2.chaoxing.com/api/login?"
	loginInfoMap = map[string]string{
		"name": config.Name,
		"account":   config.Account,
		"pwd":    config.Pwd,
		"fid":    config.Fid,
		"verify": config.Verify,
		"token": "",
		"uid":	  "",
	}
	//登录

	loginOp,err := handleLogin(loginUrl)
	fmt.Println("token = ",loginInfoMap["token"] , "uid = ",loginInfoMap["uid"])
	if err != nil {
		log.Println("登录错误：",err)
		return
	}
	coursePage := loginOp.Clone()
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
		activeOp := coursePage.Clone()
		handleActive(activeOp, activeType,activeId, "https://mobilelearn.chaoxing.com/widget/sign/pcStuSignController/preSign?activeId="+activeId+"&classId="+course.ClassId+"&fid="+loginInfoMap["fid"]+"&courseId="+course.CourseId)
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
func handleLogin(loginUrl string) (*colly.Collector,error ){
	loginOp := CreateCollector()
	err := loginOp.Post(loginUrl, map[string]string{
		"name": config.Account,
		"account":   config.Account,
		"pwd":    config.Pwd,
		"fid":    config.Fid,
		"verify": config.Verify,
		"token": "",
		"uid":	  "",
	})
	if err != nil {
		log.Println("登陆出错：", err)
		return loginOp,err
	}
	loginInfoMap["token"] = token(loginOp)
	for _,val := range loginOp.Cookies("http://i.mooc.chaoxing.com/space"){
		if val.Name == "_uid"{
			loginInfoMap["uid"] = val.Value
			break
		}
	}
	return loginOp, nil
}
func handleActive(activeOp *colly.Collector, activeType,activeId, url string) {

	if activeType == "2" {
		//签到活动
		resBody := string(normalSignin(activeOp, url).Body)
		if strings.Index(resBody,"签到成功")!=-1 {
			return
		}
		fmt.Printf("处理签到活动：",activeId)
		if strings.Index(resBody,"拍照签到")!=-1{
			fmt.Println("开始拍照签到！")
			photoUploadOp := activeOp.Clone()
			err := photoSignin(photoUploadOp,activeId)
			if err!=nil{
				log.Println("拍照签到失败！",err)
			}

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

type t struct {
	Token string `json:"_token"`
}
func token(tokenOp *colly.Collector)string{
	token := ""
	tokenOp.OnResponse(func(res *colly.Response) {
		var m t
		err := json.Unmarshal(res.Body,&m)
		if err!=nil{
			log.Println("token处理失败",err)
		}
		token = m.Token
	})
	tokenOp.Visit("https://pan-yz.chaoxing.com/api/token/uservalid")
	return token
}
type id struct {
	ObjectId string `json:"objectId"`
}
func photoUpload(photoUploadOp *colly.Collector) (string,error) {
	uploadUrl := "https://pan-yz.chaoxing.com/upload"

	fileName := "src/go_dev/chaoxing_auto_signIn/images/photo.jpg"
	file,err := os.Open(fileName)
	if err!=nil{
		return "",err
	}
	defer file.Close()
	uploadBody := &bytes.Buffer{}
	writer := multipart.NewWriter(uploadBody)
	fWriter, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return "",err
	}
	_, err = io.Copy(fWriter, file)
	if err != nil {
		return "",err
	}
	fieldMap := map[string]string{
		"puid" : loginInfoMap["uid"],
		"_token": loginInfoMap["token"],
	}
	for k, v := range fieldMap {
		_ = writer.WriteField(k, v)
	}

	err = writer.Close()
	if err != nil {
		return "",err
	}
	cookies := photoUploadOp.Cookies("https://pan-yz.chaoxing.com")
	jar,_ := cookiejar.New(nil)
	url,_ := url2.Parse("https://pan-yz.chaoxing.com")
	jar.SetCookies(url,cookies)
	client := http.Client{Jar:jar}
	res,_ := client.Post(uploadUrl,writer.FormDataContentType(),uploadBody)
	body,_ := ioutil.ReadAll(res.Body)
	var ID id
	json.Unmarshal(body,&ID)
	fmt.Println(ID.ObjectId)
	photoUploadOp.Wait()
	objectId := ID.ObjectId
	return objectId,nil
}
func photoSignin(photoSigninOp *colly.Collector, activeId string) error {
	objectId,err := photoUpload(photoSigninOp)
	url := "https://mobilelearn.chaoxing.com/pptSign/stuSignajax"
	photoSigninOp.Post(url,map[string]string{
		"name": config.Name,
		"activeId": activeId,
		"uid": loginInfoMap["uid"],
		"objectId": objectId,
	})
	fmt.Println(objectId)
	return err
}