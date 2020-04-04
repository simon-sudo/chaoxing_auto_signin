## Fxck chaoxing



1. ```powershell
   git clone https://github.com/umuoy1/chaoxing_auto_signin
   ```

2. 在`user.json`中修改个人信息，在`images`中添加你的自拍，必须为`.jpg`或`.png`格式。

3. ```powershell
   cd getCourses
   go run main.go
   ```

4. 在`user.json`中删除不需要签到课课程，`CourseId`,`ClassId`,`CourseName`，一一对应。

5. ```powershell
   go run main.go
   ```

默认每`30s`进行一次签到。

------

`user.json` 

```c
{
   "CourseId": [	//课程ID，不需要自己填
   ],
   "ClassId": [		//班级ID，不需要自己填
   ],
   "CourseName": [	//课程名，不需要自己填
   ],
   "Account": "",	//手机号
   "Pwd": "",		//密码
   "Verify": "",	//默认填 1
   "Fid": "",		//学校代号
   "Name": ""		//姓名
}
```

