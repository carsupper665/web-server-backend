// ./common/init.go
package common

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

var (
	Port                = flag.Int("port", 3000, "the listening port")
	SessionSecret       = uuid.New().String()
	CryptoSecret        = uuid.New().String()
	HMACSecret          = "HMACSecret"
	SQLitePath          = "DB.db?_busy_timeout=5000"
	LogDir              = flag.String("log-dir", "./logs", "specify the log directory")
	MemoryCacheEnabled  bool
	SyncFrequency       int
	BatchUpdateInterval int
	BatchUpdateEnabled  = false
	RelayTimeout        int
)

func LoadEnv() {

	if os.Getenv("SESSION_SECRET") != "" {
		ss := os.Getenv("SESSION_SECRET")
		if ss == "random_string" {
			log.Println("WARNING: SESSION_SECRET is set to the default value 'random_string', please change it to a random string.")
			log.Fatal("Please set SESSION_SECRET to a random string.")
		} else {
			SessionSecret = ss
		}
	}
	if os.Getenv("CRYPTO_SECRET") != "" {
		CryptoSecret = os.Getenv("CRYPTO_SECRET")
	} else {
		CryptoSecret = SessionSecret
	}

	if os.Getenv("HMAC_SECRET") != "" {
		HMACSecret = os.Getenv("HMAC_SECRET")
	} else {
		HMACSecret = SessionSecret
	}

	if os.Getenv("SQLITE_PATH") != "" {
		SQLitePath = os.Getenv("SQLITE_PATH")
	}

	if *LogDir != "" {
		var err error
		*LogDir, err = filepath.Abs(*LogDir)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := os.Stat(*LogDir); os.IsNotExist(err) {
			err = os.Mkdir(*LogDir, 0777)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	// Initialize variables from constants.go that were using environment variables
	DebugMode = os.Getenv("DEBUG") == "true"
	MemoryCacheEnabled = os.Getenv("MEMORY_CACHE_ENABLED") == "true"
	UaFilter = os.Getenv("UA_FILTER") == "true"

	// Initialize variables with GetEnvOrDefault
	SyncFrequency = GetEnvOrDefault("SYNC_FREQUENCY", 60)
	BatchUpdateInterval = GetEnvOrDefault("BATCH_UPDATE_INTERVAL", 5)
	RelayTimeout = GetEnvOrDefault("RELAY_TIMEOUT", 0)

	GlobalApiRateLimitNum = GetEnvOrDefault("GLOBAL_API_RATE_LIMIT", 60)
	GlobalApiRateLimitDuration = int64(GetEnvOrDefault("GLOBAL_API_RATE_LIMIT_DURATION", 60))
	DCWebHookUrl = GetEnvOrDefaultString("DC_WEBHOOK_URL", "")
	LatestFabricLoaderVersion = GetEnvOrDefaultString("LATEST_FABRIC_LOADER_VERSION", "")
	LatestFabricInstallerVersion = GetEnvOrDefaultString("LATEST_FABRIC_INSTALLER_VERSION", "1.1.0")
	MinecraftServerPath = GetEnvOrDefaultString("MINECRAFT_SERVER_PATH", "./minecraft_servers")

	NumPlayer = GetEnvOrDefault("NUM", 5)
	FoolChance = GetEnvOrDefault("CHANCE", 1000)

	SetUpSMTP()
	LoadVanillaServerUrls()
}

func SetUpSMTP() {
	SMTPServer = GetEnvOrDefaultString("SMTP_SERVER", "")
	SMTPPort = GetEnvOrDefault("SMTP_PORT", 587)
	SMTPSSLEnabled = GetEnvOrDefaultBool("SMTP_SSL_ENABLED", false)
	SMTPAccount = GetEnvOrDefaultString("SMTP_ACCOUNT", "")
	SMTPFrom = GetEnvOrDefaultString("SMTP_FROM", "")
	SMTPToken = GetEnvOrDefaultString("SMTP_TOKEN", "")
}

func LoadVanillaServerUrls() {
	// 讀取檔案
	data, err := os.ReadFile("common/minecraft-server-jar-downloads.json")
	if err != nil {
		log.Fatalf("無法讀取 JSON 檔案: %v", err)
	}

	// 解析 JSON 到 map
	if err := json.Unmarshal(data, &VanillaServerUrl); err != nil {
		log.Fatalf("JSON 解析失敗: %v", err)
	}
}
