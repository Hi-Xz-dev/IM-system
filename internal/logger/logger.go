package logger

import (
	"log/slog"
	"os"
)

var Log *slog.Logger //定义一个全局日志对象

func Init() {
	//日志输出到终端 格式是文本
	Log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		//Info 及以上级别会打印 Debug < Info < Warn < Error
		Level: slog.LevelInfo,
	}))
}
