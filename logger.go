package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	// 定义不同日志级别的emoji
	infoEmoji    = "ℹ️ "
	warnEmoji    = "⚠️ "
	errorEmoji   = "❌ "
	successEmoji = "✅ "
	debugEmoji   = "🔍 "

	// 预定义的翻译映射
	translations = map[string]string{
		// 错误信息翻译
		"aimeId is required":           "需要提供 aimeId 参数",
		"aimeId must be 20 characters": "aimeId 必须是20个字符",
		"AIME folder not found":        "未找到AIME文件夹",
		"Error writing aime file":      "写入aime文件时出错",
		"Warning":                      "警告",
		"not found":                    "未找到",
		"SendInput failed":             "发送输入失败",
		"Run the program with correct AIME folder path as argument": "请使用正确的AIME文件夹路径作为参数运行程序",
		"Error checking AIME folder":                                "检查AIME文件夹时出错",
		"注册服务失败喵":                                                   "注册服务失败喵",

		// 常规信息翻译
		"===== WARNING: AIME FOLDER NOT FOUND =====": "===== 警告：未找到AIME文件夹 =====",
		"The AIME folder": "AIME文件夹",
		"does not exist":  "不存在",
		"This may cause issues with the aime functionality":              "这可能会导致aime功能出现问题",
		"To resolve this issue":                                          "要解决此问题",
		"Make sure SDGA150AquaDX is correctly installed":                 "请确保正确安装了SDGA150AquaDX",
		"Run this program with the correct AIME folder path as argument": "使用正确的AIME文件夹路径作为参数运行此程序",
		"Example": "示例",
		"The program will continue, but some features may not work correctly": "程序将继续运行，但某些功能可能无法正常工作",
		"AIME folder found":                          "已找到AIME文件夹",
		"Aime ID set successfully":                   "成功设置Aime ID",
		"==========================================": "==========================================",
	}

	// 创建一个带锁的翻译器，以便于并发访问
	translatorMutex sync.RWMutex
	printer         = message.NewPrinter(language.Chinese)
)

// CustomLogger 自定义的日志记录器，支持美化和翻译
type CustomLogger struct {
	out io.Writer
}

// 初始化自定义日志记录器
func InitCustomLogger() {
	customLogger := &CustomLogger{out: os.Stdout}
	log.SetOutput(customLogger)
	log.SetFlags(log.Ldate | log.Ltime)
}

// Write 实现io.Writer接口，拦截日志输出
func (l *CustomLogger) Write(p []byte) (n int, err error) {
	msg := string(p)

	// 翻译并美化消息
	beautifiedMsg := translateAndBeautify(msg)

	// 写入到原始输出
	return l.out.Write([]byte(beautifiedMsg))
}

// 翻译并美化消息
func translateAndBeautify(msg string) string {
	translatorMutex.RLock()
	defer translatorMutex.RUnlock()

	// 确定日志级别和对应的emoji
	var emoji string

	if strings.Contains(strings.ToLower(msg), "warn") || strings.Contains(msg, "====") {
		emoji = warnEmoji
	} else if strings.Contains(strings.ToLower(msg), "error") || strings.Contains(strings.ToLower(msg), "failed") {
		emoji = errorEmoji
	} else if strings.Contains(strings.ToLower(msg), "success") {
		emoji = successEmoji
	} else if strings.Contains(strings.ToLower(msg), "found") && !strings.Contains(strings.ToLower(msg), "not found") {
		emoji = successEmoji
	} else {
		emoji = infoEmoji
	}

	// 应用翻译映射
	translatedMsg := msg
	for eng, chn := range translations {
		translatedMsg = strings.Replace(translatedMsg, eng, chn, -1)
	}

	// 添加emoji和格式化
	return fmt.Sprintf("%s %s", emoji, translatedMsg)
}

// CustomGinLogger 返回一个Gin中间件，用于美化并翻译Gin的日志
func CustomGinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()
		// 执行时间
		latency := endTime.Sub(startTime)

		// 请求方法
		reqMethod := c.Request.Method
		// 请求路由
		reqUri := c.Request.RequestURI
		// 状态码
		statusCode := c.Writer.Status()
		// 请求IP
		clientIP := c.ClientIP()

		// 状态码对应的emoji
		var statusEmoji string
		if statusCode >= 200 && statusCode < 300 {
			statusEmoji = successEmoji
		} else if statusCode >= 400 && statusCode < 500 {
			statusEmoji = warnEmoji
		} else {
			statusEmoji = errorEmoji
		}

		// 美化日志输出
		logMsg := fmt.Sprintf("%s 请求 | 状态: %d | 耗时: %v | IP: %s | %s %s",
			statusEmoji, statusCode, latency, clientIP, reqMethod, reqUri)

		fmt.Println(logMsg)
	}
}

// TranslateError 翻译错误信息
func TranslateError(err error) string {
	if err == nil {
		return ""
	}

	translatorMutex.RLock()
	defer translatorMutex.RUnlock()

	errMsg := err.Error()
	for eng, chn := range translations {
		if strings.Contains(errMsg, eng) {
			errMsg = strings.Replace(errMsg, eng, chn, -1)
		}
	}

	return fmt.Sprintf("%s %s", errorEmoji, errMsg)
}

// CustomJSONMiddleware 创建一个中间件，拦截并转换JSON响应
func CustomJSONMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建自定义ResponseWriter
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		// 如果内容类型是JSON，则添加emoji
		if strings.Contains(c.Writer.Header().Get("Content-Type"), "application/json") {
			// 将原始的JSON数据解析到map
			var objMap map[string]interface{}
			if err := json.Unmarshal(blw.body.Bytes(), &objMap); err == nil {
				// 添加emoji到不同类型的响应
				if _, hasError := objMap["error"]; hasError {
					objMap["emoji"] = errorEmoji
				} else if _, hasMessage := objMap["message"]; hasMessage {
					objMap["emoji"] = successEmoji
				} else {
					objMap["emoji"] = infoEmoji
				}

				// 将修改后的JSON写回
				newJSON, _ := json.Marshal(objMap)
				// 重设内容长度头
				c.Header("Content-Length", fmt.Sprint(len(newJSON)))
				// 写入原始的ResponseWriter
				blw.ResponseWriter.Write(newJSON)
			} else {
				// 如果解析失败，直接写回原始内容
				blw.ResponseWriter.Write(blw.body.Bytes())
			}
		} else {
			// 如果不是JSON，直接写回原始内容
			blw.ResponseWriter.Write(blw.body.Bytes())
		}
	}
}

// bodyLogWriter 自定义的ResponseWriter，用于拦截响应体
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 实现ResponseWriter的Write方法
func (w bodyLogWriter) Write(b []byte) (int, error) {
	// 只将数据写入缓冲区，不写入原始的ResponseWriter
	// 实际的写入会在中间件中完成
	return w.body.Write(b)
}

// 获取HTTP错误对应的友好中文消息
func GetHttpErrorMessage(statusCode int) string {
	messages := map[int]string{
		http.StatusBadRequest:          "请求无效",
		http.StatusUnauthorized:        "未授权访问",
		http.StatusForbidden:           "禁止访问",
		http.StatusNotFound:            "未找到资源",
		http.StatusMethodNotAllowed:    "方法不允许",
		http.StatusInternalServerError: "服务器内部错误",
		http.StatusServiceUnavailable:  "服务不可用",
	}

	if msg, ok := messages[statusCode]; ok {
		return fmt.Sprintf("%s %s", errorEmoji, msg)
	}

	return fmt.Sprintf("%s 未知错误 (状态码: %d)", errorEmoji, statusCode)
}
