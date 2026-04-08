package logger

import (
	"errors"
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger
var sugar *zap.SugaredLogger

func InitLogger() (err error) {
	logDir := "./logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return errors.New("Directory creation failed ")
	}

	// TODO: 动态读取 env 文件
	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.DebugLevel), // 日志级别
		Development: false,                                // 是否是开发模式
		Encoding:    "console",                            // 输出格式：json 或 console
		EncoderConfig: zapcore.EncoderConfig{
			// 自定义你的 EncoderConfig，或者使用 zap.NewProductionEncoderConfig()
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey, // 生产环境通常不需要函数名，减少体积
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder, // 小写编码器
			EncodeTime:     zapcore.ISO8601TimeEncoder,    // ISO8601 UTC 时间格式
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder, // 短路径编码器 (eg: /a/b/c/d.go:123)
		},
		OutputPaths:      []string{"stdout", "./" + logDir + "/app.log"}, // 输出到标准输出和文件
		ErrorOutputPaths: []string{"stderr"},                             // 错误输出
	}

	Logger, err = config.Build()
	if err != nil {
		return err
	}
	zap.ReplaceGlobals(Logger)
	sugar = Sugar()
	return nil

}

func Sugar() *zap.SugaredLogger {
	return Logger.Sugar()
}

func Info(msg string, keysAndValues ...interface{}) {
	if sugar == nil {
		log.Printf("[WARN] Logger not initialized. Msg: %s, KVs: %v", msg, keysAndValues)
		return
	}
	sugar.Infow(msg, keysAndValues...)
}

func Error(msg string, keysAndValues ...interface{}) {
	if sugar == nil {
		log.Printf("[WARN] Logger not initialized. Msg: %s, KVs: %v", msg, keysAndValues)
		return
	}
	sugar.Errorw(msg, keysAndValues...)
}

func Debug(msg string, keysAndValues ...interface{}) {
	if sugar == nil {
		log.Printf("[WARN] Logger not initialized. Msg: %s, KVs: %v", msg, keysAndValues)
		return
	}
	sugar.Debugw(msg, keysAndValues...)
}

func Warn(msg string, keysAndValues ...interface{}) {
	if sugar == nil {
		log.Printf("[WARN] Logger not initialized. Msg: %s, KVs: %v", msg, keysAndValues)
		return
	}
	sugar.Warnw(msg, keysAndValues...)
}
