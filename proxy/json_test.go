package proxy

import (
	"fmt"
	"testing"
)

func Test_Parse(t *testing.T) {
	if err := Parse("../config/servers.json", &config); err != nil {
		t.Error("解析配置文件失败!")
	} else {
		t.Log("测试通过！")
		fmt.Println("====>>>config is:", config.Dev.Chat[0])
	}
}
