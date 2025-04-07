package vars

import (
	"encoding/json"
	"os"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// a way to create a JSON configuration for wire-pod, rather than the use of env vars

var ApiConfigPath = "./apiConfig.json"

var APIConfig apiConfig

type apiConfig struct {
	Weather struct {
		Enable   bool   `json:"enable"`
		Provider string `json:"provider"`
		Key      string `json:"key"`
		Unit     string `json:"unit"`
	} `json:"weather"`
	Knowledge struct {
		Enable                 bool    `json:"enable"`
		Provider               string  `json:"provider"`
		Key                    string  `json:"key"`
		ID                     string  `json:"id"`
		Model                  string  `json:"model"`
		IntentGraph            bool    `json:"intentgraph"`
		RobotName              string  `json:"robotName"`
		OpenAIPrompt           string  `json:"openai_prompt"`
		OpenAIVoice            string  `json:"openai_voice"`
		OpenAIVoiceWithEnglish bool    `json:"openai_voice_with_english"`
		SaveChat               bool    `json:"save_chat"`
		CommandsEnable         bool    `json:"commands_enable"`
		Endpoint               string  `json:"endpoint"`
		TopP                   float32 `json:"top_p"`
		Temperature            float32 `json:"temp"`
	} `json:"knowledge"`
	STT struct {
		Service       string `json:"provider"`
		Language      string `json:"language"`
		APIKey        string `json:"api_key"`
		Endpoint      string `json:"endpoint"`
		Provider      string `json:"stt_provider"`
		Model         string `json:"model"`
	} `json:"STT"`
	Server struct {
		// false for ip, true for escape pod
		EPConfig bool   `json:"epconfig"`
		Port     string `json:"port"`
	} `json:"server"`
	HasReadFromEnv   bool `json:"hasreadfromenv"`
	PastInitialSetup bool `json:"pastinitialsetup"`
}

func WriteConfigToDisk() {
	logger.Println("Configuration changed, writing to disk")
	writeBytes, _ := json.Marshal(APIConfig)
	os.WriteFile(ApiConfigPath, writeBytes, 0644)
}

func CreateConfigFromEnv() {
	// if no config exists, create it
	if os.Getenv("WEATHERAPI_ENABLED") == "true" {
		APIConfig.Weather.Enable = true
		APIConfig.Weather.Provider = os.Getenv("WEATHERAPI_PROVIDER")
		APIConfig.Weather.Key = os.Getenv("WEATHERAPI_KEY")
		APIConfig.Weather.Unit = os.Getenv("WEATHERAPI_UNIT")
	} else {
		APIConfig.Weather.Enable = false
	}
	if os.Getenv("KNOWLEDGE_ENABLED") == "true" {
		APIConfig.Knowledge.Enable = true
		APIConfig.Knowledge.Provider = os.Getenv("KNOWLEDGE_PROVIDER")
		if os.Getenv("KNOWLEDGE_PROVIDER") == "houndify" {
			APIConfig.Knowledge.ID = os.Getenv("KNOWLEDGE_ID")
		}
		APIConfig.Knowledge.Key = os.Getenv("KNOWLEDGE_KEY")
	} else {
		APIConfig.Knowledge.Enable = false
	}
	WriteSTT()
	APIConfig.HasReadFromEnv = true
	writeBytes, _ := json.Marshal(APIConfig)
	os.WriteFile(ApiConfigPath, writeBytes, 0644)
}

func WriteSTT() {
	// was not part of the original code, so this is its own function
	// launched if stt not found in config
	APIConfig.STT.Service = os.Getenv("STT_SERVICE")
	
	// Additional configuration for Whisper service
	if os.Getenv("STT_SERVICE") == "whisper" {
		// Check for provider-specific environment variables
		if os.Getenv("WHISPER_PROVIDER") != "" {
			APIConfig.STT.Provider = os.Getenv("WHISPER_PROVIDER")
		}
		
		// API Key
		if os.Getenv("WHISPER_API_KEY") != "" {
			APIConfig.STT.APIKey = os.Getenv("WHISPER_API_KEY")
		} else if os.Getenv("OPENAI_KEY") != "" && (APIConfig.STT.Provider == "" || APIConfig.STT.Provider == "openai") {
			APIConfig.STT.APIKey = os.Getenv("OPENAI_KEY")
			APIConfig.STT.Provider = "openai"
		}
		
		// Endpoint
		if os.Getenv("WHISPER_ENDPOINT") != "" {
			APIConfig.STT.Endpoint = os.Getenv("WHISPER_ENDPOINT")
			// If endpoint is provided but no provider specified, assume it's custom
			if APIConfig.STT.Provider != "openai" && APIConfig.STT.Provider != "groq" {
				APIConfig.STT.Provider = "custom"
			}
		} else if APIConfig.STT.Provider == "openai" {
			APIConfig.STT.Endpoint = "https://api.openai.com/v1"
		} else if APIConfig.STT.Provider == "groq" {
			APIConfig.STT.Endpoint = "https://api.groq.com/openai/v1"
		}
		
		// Set language
		if os.Getenv("STT_LANGUAGE") != "" {
			APIConfig.STT.Language = os.Getenv("STT_LANGUAGE")
		}
	} else if os.Getenv("STT_SERVICE") == "whisper.cpp" {
		// For whisper.cpp, track the model being used
		if os.Getenv("WHISPER_MODEL") != "" {
			APIConfig.STT.Model = os.Getenv("WHISPER_MODEL")
		}
		
		if os.Getenv("STT_LANGUAGE") != "" {
			APIConfig.STT.Language = os.Getenv("STT_LANGUAGE")
		}
	} else if os.Getenv("STT_SERVICE") == "vosk" {
		if os.Getenv("STT_LANGUAGE") != "" {
			APIConfig.STT.Language = os.Getenv("STT_LANGUAGE")
		}
	}
}

func ReadConfig() {
	if _, err := os.Stat(ApiConfigPath); err != nil {
		CreateConfigFromEnv()
		logger.Println("API config JSON created")
	} else {
		// read config
		configBytes, err := os.ReadFile(ApiConfigPath)
		if err != nil {
			APIConfig.Knowledge.Enable = false
			APIConfig.Weather.Enable = false
			logger.Println("Failed to read API config file")
			logger.Println(err)
			return
		}
		err = json.Unmarshal(configBytes, &APIConfig)
		if err != nil {
			APIConfig.Knowledge.Enable = false
			APIConfig.Weather.Enable = false
			logger.Println("Failed to unmarshal API config JSON")
			logger.Println(err)
			return
		}
		// stt service is the only thing controlled by shell
		if APIConfig.STT.Service != os.Getenv("STT_SERVICE") {
			WriteSTT()
		}
		if !APIConfig.HasReadFromEnv {
			if APIConfig.Server.Port != os.Getenv("DDL_RPC_PORT") {
				APIConfig.HasReadFromEnv = true
				APIConfig.PastInitialSetup = true
			}
		}

		if APIConfig.Knowledge.Model == "meta-llama/Llama-2-70b-chat-hf" {
			logger.Println("Setting Together model to Llama3")
			APIConfig.Knowledge.Model = "meta-llama/Llama-3-70b-chat-hf"
		}

		writeBytes, _ := json.Marshal(APIConfig)
		os.WriteFile(ApiConfigPath, writeBytes, 0644)
		logger.Println("API config successfully read")
	}
}
