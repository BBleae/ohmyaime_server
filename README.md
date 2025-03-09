# OhMyAime 服务器

OhMyAime 是一个用于 Maimaidx 街机游戏的 Aime 卡模拟服务器，提供简单的 API 接口来设置 Aime ID 并模拟卡片刷卡操作。

## 🌟 特点

- 💳 Aime 卡模拟：无需实体卡，通过 API 模拟刷卡操作
- 🔄 状态监控：检查 Sinmai.exe 游戏进程运行状态以及 Aime 文件夹配置
- 🔎 自动发现：通过 mDNS 实现局域网内自动发现服务
- 🌏 中文友好：所有日志和错误消息支持中文显示，并带有美观的 emoji

## 📋 安装与运行

### 系统要求
- Windows 操作系统（程序使用了 Windows API 进行按键模拟）
- 已安装 SDGA150AquaDX 游戏及 AMDaemon

### 运行方法

```bash
# 使用默认 Aime 文件夹路径
./ohmyaime_server

# 使用自定义 Aime 文件夹路径
./ohmyaime_server D:\自定义路径\DEVICE
```

服务器默认监听 `0.0.0.0:8080`，可以通过浏览器或其他 HTTP 客户端访问。

## 🚀 API 接口文档

### 设置 Aime ID

用于设置 Aime ID 并触发刷卡操作。

**请求方式:** `GET`

**URL:** `/set-aime`

**参数:**

| 参数名   | 类型     | 必需 | 描述                      |
|----------|----------|------|---------------------------|
| aimeId   | string   | 是   | 20位 Aime ID 字符串       |

**成功响应:**

```json
{
  "emoji": "✅",
  "message": "成功设置 Aime ID",
  "aimeId": "01234567890123456789"
}
```

**错误响应:**

```json
{
  "emoji": "❌",
  "error": "需要提供 aimeId 参数"
}
```

或者

```json
{
  "emoji": "❌",
  "error": "aimeId 必须是20个字符"
}
```

或者

```json
{
  "emoji": "❌",
  "error": "未找到AIME文件夹: D:\\路径\\DEVICE",
  "solution": "请使用正确的AIME文件夹路径作为参数运行程序"
}
```

### 获取状态信息

用于检查服务和配置状态。

**请求方式:** `GET`

**URL:** `/status`

**参数:** 无

**成功响应:**

```json
{
  "emoji": "ℹ️",
  "sinmai_running": true,
  "aime_folder_path": "D:\\SDGA150AquaDX\\AMDaemon\\DEVICE",
  "aime_folder_exists": true
}
```

**错误情况响应:**

```json
{
  "emoji": "⚠️",
  "sinmai_running": false,
  "aime_folder_path": "D:\\SDGA150AquaDX\\AMDaemon\\DEVICE",
  "aime_folder_exists": false,
  "aime_folder_error": "未找到AIME文件夹: D:\\SDGA150AquaDX\\AMDaemon\\DEVICE",
  "solution": "请使用正确的AIME文件夹路径作为参数运行程序"
}
```

## 📝 工作原理

1. 服务器启动时会检查配置的 AIME 文件夹路径是否存在
2. 当收到 `/set-aime` 请求时，服务器会:
   - 将提供的 Aime ID 写入 `aime.txt` 文件
   - 模拟按下回车键以触发游戏中的刷卡识别
3. 通过 `/status` 接口可随时检查服务状态

## 🔧 故障排除

如果遇到问题:

1. 确保 SDGA150AquaDX 正确安装
2. 确认 Aime 文件夹路径正确，如有必要请指定自定义路径
3. 检查 Sinmai.exe 是否正在运行
4. 通过 `/status` 接口查看详细的配置状态

## 📜 许可证

本项目使用 MIT 许可证。