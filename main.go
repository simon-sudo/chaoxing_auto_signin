package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	url2 "net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

var userInfoMap = make(map[string]string)
var isExecuted = make(map[string]bool)
type Config struct {
	CourseId []string 	`json:"CourseId"`
	ClassId []string	`json:"ClassId"`
	CourseName []string `json:"CourseName"`
	Account string		`json:"Account"`
	Pwd string			`json:"Pwd"`
	Verify string		`json:"Verify"`
	Fid string			`json:"Fid"`
	Name string			`json:"Name"`
}
var conf Config
func init() {
	//configPath := "src/go_dev/chaoxing_auto_signIn/config/user.json"
	configPath,_ :=os.Getwd()
	configPath += "/config/user.json"
	data,err := ioutil.ReadFile(configPath)
	if err!=nil{
		log.Println(err)
	}
	err = json.Unmarshal(data,&conf)
	if err!=nil{
		log.Println(err)
	}
}
func main() {
	//登录页
	const loginUrl = "http://passport2.chaoxing.com/api/login?"
	userInfoMap = map[string]string{
		"name":    conf.Name,
		"account": conf.Account,
		"pwd":     conf.Pwd,
		"fid":     conf.Fid,
		"verify":  conf.Verify,
		"getToken":   "",
		"uid":     "",
	}

	//登录

	loginOp, err := handleLogin(loginUrl)
	if err != nil {
		log.Println("登录失败：", err)
		fmt.Scanln()
		return
	}
	fmt.Println("登录成功！")
	fmt.Println("Fxcking chaoxing")
	for {
		//处理课程活动页面
		for ind := range conf.ClassId {
			coursePage := loginOp.Clone()
			handleCourse(coursePage, ind)
		}
		time.Sleep(time.Second*120)
	}
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
type result struct {
	Result bool `json:"result"`
	ErrorMsg string `json:"errorMsg"`
}
var res result
func handleLogin(loginUrl string) (*colly.Collector, error) {
	loginOp := CreateCollector()
	loginOp.OnResponse(func(resb *colly.Response) {
		json.Unmarshal(resb.Body,&res)
	})
	err := loginOp.Post(loginUrl, map[string]string{
		"name":    conf.Account,
		"pwd":     conf.Pwd,
		"fid":     conf.Fid,
		"verify":  conf.Verify,
	})
	if res.Result == false{
		return nil,fmt.Errorf(res.ErrorMsg)
	}
	if err != nil {
		return nil, err
	}
	userInfoMap["getToken"] = getToken(loginOp)
	for _, val := range loginOp.Cookies("http://i.mooc.chaoxing.com/space") {
		if val.Name == "_uid" {
			userInfoMap["uid"] = val.Value
			break
		}
	}
	return loginOp, nil
}
func handleCourse(coursePage *colly.Collector,ind int) {
	coursePage.OnHTML("#startList .Mct", func(element *colly.HTMLElement) {
		s := element.Attr("onclick")
		reg1 := regexp.MustCompile(`\(\d+`)
		reg2 := regexp.MustCompile(`,\d+,`)
		activeId := reg1.FindString(s)
		activeType := reg2.FindString(s)
		activeId = activeId[1:]
		activeType = activeType[1 : len(activeType)-1]
		activeOp := coursePage.Clone()
		handleActive(activeOp, ind,activeType, activeId, "https://mobilelearn.chaoxing.com/widget/sign/pcStuSignController/preSign?activeId="+activeId+"&classId="+conf.ClassId[ind]+"&fid="+userInfoMap["fid"]+"&courseId="+conf.CourseId[ind])
	})

	_ = coursePage.Visit("https://mobilelearn.chaoxing.com/widget/pcpick/stu/index?courseId=" + conf.CourseId[ind] + "&jclassId=" + conf.ClassId[ind])
}
func handleActive(activeOp *colly.Collector,ind int, activeType, activeId, url string) {

	if isExecuted[activeId]{
		return
	}
	isExecuted[activeId]=true
	if activeType == "2" {
		//签到活动
		handleSignin(activeOp, ind,activeId, url)
	}
}
func handleSignin(activeOp *colly.Collector, ind int,activeId, url string) {

	fmt.Println("处理签到活动：", conf.CourseName[ind])
	resBody := string(normalSignin(activeOp, url).Body)
	if strings.Index(resBody, "签到成功") != -1 {
		fmt.Println("签到成功！")
		return
	}
	if strings.Index(resBody, "拍照签到") != -1 {
		//fmt.Println("开始拍照签到！")
		photoUploadOp := activeOp.Clone()
		resBody = photoSignin(photoUploadOp, activeId)
		if strings.Index(resBody, "success") != -1 {
			fmt.Println("拍照签到成功！")
		} else {
			fmt.Println("拍照签到失败！")
		}

	} else if strings.Index(resBody, "手势签到") != -1 {
		//fmt.Println("开始手势签到！")
		gestureSigninOp := activeOp.Clone()
		resBody = gestureSignin(gestureSigninOp,activeId)
		if strings.Index(resBody, "success") != -1 {
			fmt.Println("手势签到成功！")
		} else {
			fmt.Println("手势签到失败！")
		}
	}
}
func normalSignin(activeOp *colly.Collector, url string) *colly.Response {
	var res *colly.Response
	activeOp.OnResponse(func(resp *colly.Response) {
		res = resp
	})
	_ = activeOp.Visit(url)
	return res
}

func photoSignin(photoSigninOp *colly.Collector, activeId string) string {
	objectId, err := photoUpload(photoSigninOp)
	if err != nil {
		log.Println(objectId, err)
		return ""
	}
	url := "https://mobilelearn.chaoxing.com/pptSign/stuSignajax"
	var resBody string
	photoSigninOp.OnResponse(func(res *colly.Response) {
		resBody = string(res.Body)
	})
	err = photoSigninOp.Post(url, map[string]string{
		"name":     conf.Name,
		"activeId": activeId,
		"uid":      userInfoMap["uid"],
		"objectId": objectId,
	})
	return resBody
}
func gestureSignin(gestureSigninOp *colly.Collector, activeId string)string {
	url := "https://mobilelearn.chaoxing.com/pptSign/stuSignajax"
	var resBody string
	gestureSigninOp.OnResponse(func(res *colly.Response) {
		resBody = string(res.Body)
	})
	_ = gestureSigninOp.Post(url,map[string]string{
		"name":     conf.Name,
		"activeId": activeId,
		"uid":      userInfoMap["uid"],
	})
	return resBody
}
func getToken(tokenOp *colly.Collector) string {
	type t struct {
		Token string `json:"_token"`
	}
	tokenStr := ""
	tokenOp.OnResponse(func(res *colly.Response) {
		var m t
		err := json.Unmarshal(res.Body, &m)
		if err != nil {
			panic("token处理错误")
		}
		tokenStr = m.Token
	})
	_ = tokenOp.Visit("https://pan-yz.chaoxing.com/api/getToken/uservalid")
	return tokenStr
}
func photoUpload(photoUploadOp *colly.Collector) (string, error) {
	uploadUrl := "https://pan-yz.chaoxing.com/upload"
	fileName := ""
	imgPath,_ := os.Getwd()
	imgPath += "/images/"
	imgFiles,_ := ioutil.ReadDir(imgPath)
	i := rand.Int()%len(imgFiles)
	fileName = imgPath+imgFiles[i].Name()
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()
	uploadBody := &bytes.Buffer{}
	writer := multipart.NewWriter(uploadBody)
	fWriter, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(fWriter, file)
	if err != nil {
		return "", err
	}
	fieldMap := map[string]string{
		"puid":   userInfoMap["uid"],
		"_token": userInfoMap["getToken"],
	}
	for k, v := range fieldMap {
		_ = writer.WriteField(k, v)
	}

	err = writer.Close()
	if err != nil {
		return "", err
	}
	cookies := photoUploadOp.Cookies("https://pan-yz.chaoxing.com")
	jar, _ := cookiejar.New(nil)
	url, _ := url2.Parse("https://pan-yz.chaoxing.com")
	jar.SetCookies(url, cookies)
	client := http.Client{Jar: jar}
	res, _ := client.Post(uploadUrl, writer.FormDataContentType(), uploadBody)
	body, _ := ioutil.ReadAll(res.Body)
	type uploadRes struct {
		ObjectId string `json:"objectId"`
		Result   bool   `json"result"`
	}
	var resData uploadRes
	_ = json.Unmarshal(body, &resData)
	if resData.Result == false {
		return "图片上传失败", err
	}
	objectId := resData.ObjectId
	return objectId, nil
}
