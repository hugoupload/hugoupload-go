# hugo 分片上传包

```
go get github.com/hyahm/hugoPartUpload
```
- 需要下载最新版go

```go
package main

import (
	"log"
	"github.com/hyahm/hugoPartUpload"
)

func main() {
	pc := hugoPartUpload.PartClient{
		Filename:    "C:\\Users\\Admin\\Desktop\\a.mp4",  //本地视频路径
		Token:       "xxxxxxxxxxxxxx",   // 后台里面
		Identifier:  "aaaa",  // 后台里面
		User:        "test",  // 后台里面
		Title:       "test",  // 后台里面
		Rule:        "test",  // 后台里面
		Cat:         "mm_手机下载",  // 后台里面
		Subcat:      []string{"经典"},  // 后台里面
		Actor:       "test",  // 后台里面
		Domain:      "http://192.168.50.72",   // 上传接口的域名
	}
	err := pc.PartUpload()
	if err != nil {
		log.Fatal(err)
	}
}
```

> 国内需要设置代理
```
$env:GOPORXY="https://goproxy.cn"  // windows
export GOPORXY="https://goproxy.cn"  // 非windows
```

> 运行
```
go run exemple/main.go
```