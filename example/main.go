package main

import (
	"fmt"
	"time"

	"github.com/hyahm/golog"
	"github.com/hyahm/hugoPartUpload"
)

func main() {
	defer golog.Sync()
	start := time.Now()
	pc := hugoPartUpload.PartClient{
		// 上传文件的路径
		Filename: "aaa.mp4",
		// token
		Token: "xxxxxxxxxxxx",
		// 上传的用户  user 还为了验证 token
		User: "admin",
		// 唯一标识 随机生成
		Identifier: "3333333333",
		// 视频的标题
		Title: "aafdfdfdf",
		// 使用的规则， 必须是user 所属用户的
		Rule: "videotest",
		// 使用的分类， 必须是user 所属用户的
		Cat: "test",
		// 使用的子分类， 必须是Cat的子分类
		Subcat: []string{"网红热点"},
		// 可不填
		Actor: "test",
		// 存入服务器的文件名  建议是identifer为前缀名， 注意后缀， 后缀名必须是浏览器支持的播放的格式
		// Domain:      "http://127.0.0.1:8080",
	}
	err := pc.PartUpload()
	if err != nil {
		golog.Fatal(err)
	}
	fmt.Println(time.Since(start).Seconds())
}
