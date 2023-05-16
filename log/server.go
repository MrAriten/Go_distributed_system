package log

import (
	"io"
	stlog "log"
	"net/http"
	"os"
)

var log *stlog.Logger

type fileLog string //日志的地址

// 将post请求写入到log中，返回值是写入的数据量以及可能发生的错误
// fileLog 是一个自定义的类型别名，它本质上是一个字符串类型。
// 然而，Go语言对于类型别名并不具有隐式的类型转换能力。
// 因此，在这里使用 string(fl) 的方式将 fl 转换为 string 类型，以便在 os.OpenFile 函数中作为文件路径参数传递。
// os.O_CREATE：如果文件不存在，创建文件。如果文件存在，不执行任何操作。
// os.O_WRONLY：以只写模式打开文件。
// os.O_APPEND：将数据追加到文件末尾，而不是覆盖文件中的内容。
// 0600 是指定文件的权限，它是一个八进制数。在这里，它表示文件的权限位为 0600，即读写权限属于所有者，没有执行权限。这是一种常见的文件权限设置，表示只有文件的所有者有读写权限，其他用户没有访问权限。
func (fl fileLog) Write(data []byte) (int, error) {
	f, err := os.OpenFile(string(fl), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return f.Write(data)
}

// New返回一个*Logger类型
// New的第一个参数是io.Writer接口，只要实现了Write函数即可
// const LstdFlags int = Ldate | Ltime
// LstdFlags initial values for the standard logger —— 能够初始化日志的日期和时间
func Run(destination string) {
	log = stlog.New(fileLog(destination), "[go] - ", stlog.LstdFlags)
}

// 注册响应处理函数
func RegisterHandlers() {
	http.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) { //匿名函数直接作为参数
		switch r.Method {
		case http.MethodPost: //如果是POST请求
			msg, err := io.ReadAll(r.Body)
			if err != nil || len(msg) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			write(string(msg))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed) //其他方法未定义
			return
		}
	})
}

func write(message string) {
	log.Printf("%v\n", message) //logger会自动将日志输出到文件中
}
