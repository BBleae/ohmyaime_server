package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grandcat/zeroconf"
)

var AIME_FOLDER_PATH string

func init() {
	// 初始化自定义日志记录器
	InitCustomLogger()
	AIME_FOLDER_PATH = getAimeFolder()
	// Check if AIME folder exists
	if _, err := os.Stat(AIME_FOLDER_PATH); os.IsNotExist(err) {
		log.Println("===== WARNING: AIME FOLDER NOT FOUND =====")
		log.Printf("The AIME folder '%s' does not exist.\n", AIME_FOLDER_PATH)
		log.Println("This may cause issues with the aime functionality.")
		log.Println("")
		log.Println("To resolve this issue:")
		log.Println("1. Make sure SDGA150AquaDX is correctly installed")
		log.Println("2. Run this program with the correct AIME folder path as argument:")
		log.Println("   Example: ./ohmyaime D:\\path\\to\\DEVICE")
		log.Println("")
		log.Println("The program will continue, but some features may not work correctly.")
		log.Println("==========================================")
	} else {
		log.Printf("AIME folder found: %s\n", AIME_FOLDER_PATH)
	}
}

func getAimeFolder() string {
	defaultFolder := "D:\\SDGA150AquaDX\\AMDaemon\\DEVICE"
	if CheckFolderExists(defaultFolder) {
		return defaultFolder
	}
	//read from args
	if len(os.Args) > 1 {
		if CheckFolderExists(os.Args[1]) {
			return os.Args[1]
		}
	}
	// If we reached here, none of the folders exist
	// We'll return the default folder anyway, but the program will show warnings
	return defaultFolder
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
