package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grandcat/zeroconf"
)

var AIME_FOLDER_PATH string

func init() {
	// 初始化自定义日志记录器
	InitCustomLogger()

	// 获取AIME文件夹路径
	// getAimeFolder函数会自动处理错误情况并在必要时panic
	AIME_FOLDER_PATH = getAimeFolder()

	// 由于成功找到了AIME文件夹（否则会在getAimeFolder中panic），输出确认信息
	log.Printf("AIME folder found: %s\n", AIME_FOLDER_PATH)
}

func getAimeFolder() string {
	// 1. 首先优先检查命令行参数
	if len(os.Args) > 1 {
		if CheckFolderExists(os.Args[1]) {
			log.Printf("使用命令行指定的AIME文件夹: %s", os.Args[1])
			return os.Args[1]
		} else {
			log.Printf("命令行参数指定的文件夹不存在: %s", os.Args[1])
		}
	}

	// 2. 检查Sinmai.exe是否正在运行，并获取其路径
	sinmaiPath, found := GetProcessPath("Sinmai.exe")
	if !found {
		log.Printf("未检测到Sinmai.exe正在运行，无法自动定位AIME文件夹")
		panic("无法找到AIME文件夹。请确保Sinmai.exe正在运行，或通过命令行参数指定AIME文件夹路径。例如: ./ohmyaime D:\\SDGA-1.50\\AMDaemon\\DEVICE")
	}

	log.Printf("找到Sinmai.exe路径: %s", sinmaiPath)

	// 获取Sinmai.exe所在的目录
	sinmaiDir := filepath.Dir(sinmaiPath)

	// 3. 从Sinmai.exe所在目录开始，向上递归查找
	currentPath := sinmaiDir

	// 定义可能的文件夹名称
	folderName := "AMDaemon\\DEVICE"

	// 存储已检查的路径，避免重复检查
	checkedPaths := make(map[string]bool)

	for {
		// 构建可能的AIME文件夹路径
		possiblePath := filepath.Join(currentPath, folderName)

		// 如果已经检查过，跳过
		if _, checked := checkedPaths[possiblePath]; checked {
			break
		}
		checkedPaths[possiblePath] = true

		log.Printf("检查潜在的AIME文件夹: %s", possiblePath)

		// 检查文件夹是否存在
		if CheckFolderExists(possiblePath) {
			log.Printf("找到AIME文件夹: %s", possiblePath)
			return possiblePath
		}

		// 获取上一级目录
		parentDir := filepath.Dir(currentPath)

		// 如果到达了根目录，或者发现我们不再有上升的空间（即当前路径与父目录相同）
		if parentDir == currentPath {
			break
		}

		// 上升到父目录继续查找
		currentPath = parentDir
	}

	// 如果到这里还没找到，抛出详细的错误
	errorMsg := "无法找到AIME文件夹。自动查找已检查以下路径:\n"
	for path := range checkedPaths {
		errorMsg += fmt.Sprintf("  - %s\n", path)
	}
	errorMsg += "\n可能的解决方法:\n"
	errorMsg += "1. 确保Sinmai.exe正在运行\n"
	errorMsg += "2. 在命令行中指定AIME文件夹的完整路径: ./ohmyaime 完整路径\n"
	errorMsg += "3. 手动创建AIME文件夹: " + filepath.Join(sinmaiDir, folderName)

	panic(errorMsg)
}

func CheckFolderExists(folder string) bool {
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		log.Printf("Warning: %s not found", folder)
		return false
	}
	return true
}

func WriteAimeIdToFile(aimeId string) {
	file, err := os.Create(AIME_FOLDER_PATH + "\\aime.txt")
	if err != nil {
		log.Printf("Error writing aime file: %v", err)
		return
	}
	defer file.Close()
	if _, err := file.WriteString(aimeId); err != nil {
		log.Printf("Error writing content to aime file: %v", err)
	}
}

func serveMDNS() {
	server, err := zeroconf.Register(
		"ohmyaime",                         // 实例名，最终将解析为 ohmyaime.local
		"_http._tcp",                       // 服务类型，可以根据需要更改喵
		"local.",                           // 域名固定为 local.
		8080,                               // 设备端口，可以按需调整
		[]string{"txtv=0", "lo=1", "la=2"}, // 服务属性，可以按需调整
		nil,
	)
	if err != nil {
		log.Fatal("注册服务失败喵: ", err)
	}
	defer server.Shutdown()
}

const LISTEN = "0.0.0.0:8080"

func main() {
	go serveMDNS()
	// 使用自定义的Gin配置
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	// 添加自定义日志中间件
	r.Use(CustomGinLogger())
	r.Use(gin.Recovery())
	r.Use(CustomJSONMiddleware())
	r.GET("/set-aime", func(c *gin.Context) {
		aimeId := c.Query("aimeId")
		if aimeId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "aimeId is required"})
			return
		}
		if len(aimeId) != 20 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "aimeId must be 20 characters"})
			return
		}
		// Check if AIME folder exists before writing
		if _, err := os.Stat(AIME_FOLDER_PATH); os.IsNotExist(err) {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":    fmt.Sprintf("AIME folder not found: %s", AIME_FOLDER_PATH),
				"solution": "Run the program with correct AIME folder path as argument",
			})
			return
		}
		WriteAimeIdToFile(aimeId)
		sendGlobalEnterKey("enter", 500*time.Millisecond)
		c.JSON(http.StatusOK, gin.H{
			"message": "Aime ID set successfully",
			"aimeId":  aimeId,
		})
	})
	r.GET("/status", func(c *gin.Context) {
		// Check if Sinmai.exe is running
		sinmaiRunning := IsProcessRunning("Sinmai.exe")
		// Check if AIME folder exists
		aimeFolderExists := false
		var aimeFolderError string
		var solution string
		if _, err := os.Stat(AIME_FOLDER_PATH); os.IsNotExist(err) {
			aimeFolderError = fmt.Sprintf("AIME folder not found: %s", AIME_FOLDER_PATH)
			solution = "Run the program with correct AIME folder path as argument"
		} else if err != nil {
			aimeFolderError = fmt.Sprintf("Error checking AIME folder: %v", err)
		} else {
			aimeFolderExists = true
		}
		// Build response
		response := gin.H{
			"sinmai_running":     sinmaiRunning,
			"aime_folder_path":   AIME_FOLDER_PATH,
			"aime_folder_exists": aimeFolderExists,
		}
		if aimeFolderError != "" {
			response["aime_folder_error"] = aimeFolderError
			if solution != "" {
				response["solution"] = solution
			}
		}
		c.JSON(http.StatusOK, response)
	})
	log.Printf("🚀 服务器启动成功，监听端口 %s", LISTEN)
	r.Run(LISTEN)
}
