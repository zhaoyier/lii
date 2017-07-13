package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

func init() {

}

//将json config文件转换成json对象
func Parse(filename string, structure interface{}) error {
	file, err := os.Open(filename) // For read access.
	if err != nil {
		fmt.Println("加载" + filename + "配置文件出错")
		return errors.New("加载" + filename + "配置文件出错...")
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println("加载" + filename + "配置文件出错")
		return errors.New("加载" + filename + "配置文件出错")
	}
	return json.Unmarshal(data, &structure)
}

func Process(ft int32, msg []byte, fn func([]byte)) {
	//查找连接句柄
	raw := GetRaw(ft, 1) //TODO
	raw.conn.Write(msg)
	raw.conn.Read(msg)
	fn(msg)
}

func GetRaw(ft int32, seqID int64) *Connect {
	for _, v := range clients {
		if ft >= v.min && ft <= v.max {
			if len(v.conns) == 1 {
				return v.conns[0]
			} else if len(v.conns) > 1 {
				return v.conns[seqID%int64(len(v.conns))]
			}
			return nil
		}
	}
	return nil
}
