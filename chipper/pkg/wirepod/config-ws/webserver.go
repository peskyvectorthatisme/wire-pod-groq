package webserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/scripting"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"github.com/kercre123/wire-pod/chipper/pkg/wirepod/localization"
	processreqs "github.com/kercre123/wire-pod/chipper/pkg/wirepod/preqs"
	botsetup "github.com/kercre123/wire-pod/chipper/pkg/wirepod/setup"
)

var SttInitFunc func() error

func apiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	switch strings.TrimPrefix(r.URL.Path, "/api/") {
	case "add_custom_intent":
		handleAddCustomIntent(w, r)
	case "edit_custom_intent":
		handleEditCustomIntent(w, r)
	case "get_custom_intents_json":
		handleGetCustomIntentsJson(w)
	case "remove_custom_intent":
		handleRemoveCustomIntent(w, r)
	case "is_api_v3":
		fmt.Fprintf(w, "it is!")
	case "get_config":
		handleGetConfig(w)
	case "set_weather_api":
		handleSetWeatherAPI(w, r)
	case "get_weather_api":
		handleGetWeatherAPI(w)
	case "set_kg_api":
		handleSetKGAPI(w, r)
	case "get_kg_api":
		handleGetKGAPI(w)
	case "set_stt_info":
		handleSetSTTInfo(w, r)
	case "get_stt_info":
		handleGetSTTInfo(w)
	case "get_download_status":
		handleGetDownloadStatus(w)
	case "get_logs":
		handleGetLogs(w)
	case "get_debug_logs":
		handleGetDebugLogs(w)
	case "is_running":
		handleIsRunning(w)
	case "delete_chats":
		handleDeleteChats(w)
	case "get_ota":
		handleGetOTA(w, r)
	case "get_version_info":
		handleGetVersionInfo(w)
	case "generate_certs":
		handleGenerateCerts(w)
	case "test_endpoint":
		handleTestEndpoint(w, r)
	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

func handleAddCustomIntent(w http.ResponseWriter, r *http.Request) {
	var intent vars.CustomIntent
	if err := json.NewDecoder(r.Body).Decode(&intent); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if anyEmpty(intent.Name, intent.Description, intent.Intent) || len(intent.Utterances) == 0 {
		http.Error(w, "missing required field (name, description, utterances, and intent are required)", http.StatusBadRequest)
		return
	}
	intent.LuaScript = strings.TrimSpace(intent.LuaScript)
	if intent.LuaScript != "" {
		if err := scripting.ValidateLuaScript(intent.LuaScript); err != nil {
			http.Error(w, "lua validation error: "+err.Error(), http.StatusBadRequest)
			return
		}
	}
	vars.CustomIntentsExist = true
	vars.CustomIntents = append(vars.CustomIntents, intent)
	saveCustomIntents()
	fmt.Fprint(w, "Intent added successfully.")
}

func handleEditCustomIntent(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Number int `json:"number"`
		vars.CustomIntent
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if request.Number < 1 || request.Number > len(vars.CustomIntents) {
		http.Error(w, "invalid intent number", http.StatusBadRequest)
		return
	}
	intent := &vars.CustomIntents[request.Number-1]
	if request.Name != "" {
		intent.Name = request.Name
	}
	if request.Description != "" {
		intent.Description = request.Description
	}
	if len(request.Utterances) != 0 {
		intent.Utterances = request.Utterances
	}
	if request.Intent != "" {
		intent.Intent = request.Intent
	}
	if request.Params.ParamName != "" {
		intent.Params.ParamName = request.Params.ParamName
	}
	if request.Params.ParamValue != "" {
		intent.Params.ParamValue = request.Params.ParamValue
	}
	if request.Exec != "" {
		intent.Exec = request.Exec
	}
	if request.LuaScript != "" {
		intent.LuaScript = request.LuaScript
		if err := scripting.ValidateLuaScript(intent.LuaScript); err != nil {
			http.Error(w, "lua validation error: "+err.Error(), http.StatusBadRequest)
			return
		}
	}
	if len(request.ExecArgs) != 0 {
		intent.ExecArgs = request.ExecArgs
	}
	intent.IsSystemIntent = false
	saveCustomIntents()
	fmt.Fprint(w, "Intent edited successfully.")
}

func handleGetCustomIntentsJson(w http.ResponseWriter) {
	if !vars.CustomIntentsExist {
		http.Error(w, "you must create an intent first", http.StatusBadRequest)
		return
	}
	customIntentJSONFile, err := os.ReadFile(vars.CustomIntentsPath)
	if err != nil {
		http.Error(w, "could not read custom intents file", http.StatusInternalServerError)
		logger.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(customIntentJSONFile)
}

func handleRemoveCustomIntent(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Number int `json:"number"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if request.Number < 1 || request.Number > len(vars.CustomIntents) {
		http.Error(w, "invalid intent number", http.StatusBadRequest)
		return
	}
	vars.CustomIntents = append(vars.CustomIntents[:request.Number-1], vars.CustomIntents[request.Number:]...)
	saveCustomIntents()
	fmt.Fprint(w, "Intent removed successfully.")
}

func handleSetWeatherAPI(w http.ResponseWriter, r *http.Request) {
	var config struct {
		Provider string `json:"provider"`
		Key      string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if config.Provider == "" {
		vars.APIConfig.Weather.Enable = false
	} else {
		vars.APIConfig.Weather.Enable = true
		vars.APIConfig.Weather.Key = strings.TrimSpace(config.Key)
		vars.APIConfig.Weather.Provider = config.Provider
	}
	vars.WriteConfigToDisk()
	fmt.Fprint(w, "Changes successfully applied.")
}

func handleGetWeatherAPI(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vars.APIConfig.Weather)
}

func handleSetKGAPI(w http.ResponseWriter, r *http.Request) {
	if err := json.NewDecoder(r.Body).Decode(&vars.APIConfig.Knowledge); err != nil {
		fmt.Println(err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	vars.WriteConfigToDisk()
	fmt.Fprint(w, "Changes successfully applied.")
}

func handleGetKGAPI(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vars.APIConfig.Knowledge)
}

func handleSetSTTInfo(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Language    string `json:"language"`      // General language setting
		Provider    string `json:"provider"`      // Main STT Service (vosk, whisper, whisper.cpp)
		SttProvider string `json:"stt_provider"` // Specific Whisper provider (openai, groq, custom)
		APIKey      string `json:"api_key"`      // Whisper API Key
		Endpoint    string `json:"endpoint"`      // Whisper Custom Endpoint
		Model       string `json:"model"`       // Whisper.cpp Model
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Keep track if a restart is needed
	needsRestart := false
	madeChanges := false

	// Store current values before potential changes to check for restart necessity
	currentService := vars.APIConfig.STT.Service
	currentLanguage := vars.APIConfig.STT.Language
	currentWhisperProvider := vars.APIConfig.STT.Provider
	currentAPIKey := vars.APIConfig.STT.APIKey
	currentEndpoint := vars.APIConfig.STT.Endpoint
	currentModel := vars.APIConfig.STT.Model

	// Check if the main STT service/provider needs to change
	if request.Provider != "" && request.Provider != currentService {
		needsRestart = true
		madeChanges = true
		vars.APIConfig.STT.Service = request.Provider
	}

	// Use the potentially updated service for the rest of the function
	service := vars.APIConfig.STT.Service

	if service == "vosk" {
		if request.Language != "" {
			if !isValidLanguage(request.Language, localization.ValidVoskModels) {
				http.Error(w, "language not valid for Vosk", http.StatusBadRequest)
				return
			}
			if request.Language != currentLanguage {
				needsRestart = true
				madeChanges = true
				vars.APIConfig.STT.Language = request.Language
			}
			if !isDownloadedLanguage(vars.APIConfig.STT.Language, vars.DownloadedVoskModels) {
				go localization.DownloadVoskModel(vars.APIConfig.STT.Language)
				// Don't save config yet, let download complete
				fmt.Fprint(w, "downloading")
				return
			}
		}
	} else if service == "whisper.cpp" {
		// Handle whisper.cpp model
		if request.Model != "" && request.Model != currentModel {
			needsRestart = true
			madeChanges = true
			vars.APIConfig.STT.Model = request.Model
		}
		// Handle language for whisper.cpp (just store it, doesn't require restart)
		if request.Language != "" && request.Language != currentLanguage {
			madeChanges = true
			vars.APIConfig.STT.Language = request.Language
		}
	} else if service == "whisper" {
		// For Whisper API, none of these changes require a full restart
		// We'll update values and reinitialize the service using ReloadVosk()
		
		// Handle Whisper API settings
		whisperSpecificProvider := vars.APIConfig.STT.Provider // Default to current
		if request.SttProvider != "" && request.SttProvider != currentWhisperProvider {
			madeChanges = true
			whisperSpecificProvider = request.SttProvider
			vars.APIConfig.STT.Provider = whisperSpecificProvider // Update the specific provider
		}

		if request.APIKey != "" && request.APIKey != currentAPIKey {
			madeChanges = true
			vars.APIConfig.STT.APIKey = request.APIKey
		}

		// Determine the correct endpoint
		newEndpoint := currentEndpoint // Default to current
		if whisperSpecificProvider == "openai" {
			newEndpoint = "https://api.openai.com/v1"
		} else if whisperSpecificProvider == "groq" {
			newEndpoint = "https://api.groq.com/openai/v1"
		} else if whisperSpecificProvider == "custom" && request.Endpoint != "" {
			newEndpoint = request.Endpoint
		}

		if newEndpoint != currentEndpoint {
			madeChanges = true
			vars.APIConfig.STT.Endpoint = newEndpoint
		}

		// Language for Whisper API - just store it, doesn't affect operation
		if request.Language != "" && request.Language != currentLanguage {
			madeChanges = true
			vars.APIConfig.STT.Language = request.Language
		}
	} else if request.Provider != "" {
		// This case handles when a Provider was sent in the request, but it's not one of the known ones
		http.Error(w, "service must be vosk, whisper.cpp, or whisper", http.StatusBadRequest)
		return
	}

	// If no changes were made, just return success
	if !madeChanges {
		fmt.Fprint(w, "No changes detected.")
		return
	}

	// Save config changes
	vars.APIConfig.PastInitialSetup = true
	vars.WriteConfigToDisk()
	logger.Println("Configuration changed, writing to disk")

	if needsRestart {
		logger.Println("STT service or critical settings changed. A restart is needed for changes to take effect.")
		fmt.Fprint(w, "Settings saved successfully. Please restart the server for the changes to take effect.")
	} else {
		// For Whisper API settings, we can try to reload without a restart
		if service == "whisper" {
			processreqs.ReloadVosk()
			logger.Println("Reloaded Whisper API settings")
			fmt.Fprint(w, "Settings saved and applied successfully.")
		} else {
			fmt.Fprint(w, "Settings saved successfully.")
		}
	}
}

func handleGetDownloadStatus(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(localization.DownloadStatus))
	if localization.DownloadStatus == "success" || strings.Contains(localization.DownloadStatus, "error") {
		localization.DownloadStatus = "not downloading"
	}
}

func handleGetSTTInfo(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vars.APIConfig.STT)
}

func handleGetConfig(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vars.APIConfig)
}

func handleGetLogs(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(logger.LogList))
}

func handleGetDebugLogs(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(logger.LogTrayList))
}

func handleIsRunning(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("true"))
}

func handleDeleteChats(w http.ResponseWriter) {
	vars.RememberedChats = []vars.RememberedChat{}
	fmt.Fprint(w, "done")
}

func handleGetOTA(w http.ResponseWriter, r *http.Request) {
	otaName := strings.Split(r.URL.Path, "/")[3]
	targetURL, err := url.Parse("https://archive.org/download/vector-pod-firmware/" + strings.TrimSpace(otaName))
	if err != nil {
		http.Error(w, "failed to parse URL", http.StatusInternalServerError)
		return
	}
	req, err := http.NewRequest(r.Method, targetURL.String(), nil)
	if err != nil {
		http.Error(w, "failed to create request", http.StatusInternalServerError)
		return
	}
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "failed to perform request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		http.Error(w, "failed to copy response body", http.StatusInternalServerError)
	}
}

func handleGetVersionInfo(w http.ResponseWriter) {
	var installedVer string
	ver, err := os.ReadFile(vars.VersionFile)
	if err == nil {
		installedVer = strings.TrimSpace(string(ver))
	}
	currentVer, err := GetLatestReleaseTag("kercre123", "WirePod")
	if err != nil {
		http.Error(w, "error communicating with github (ver): "+err.Error(), http.StatusInternalServerError)
		return
	}
	currentCommit, err := GetLatestCommitSha()
	if err != nil {
		http.Error(w, "error communicating with github (commit): "+err.Error(), http.StatusInternalServerError)
		return
	}
	type VersionInfo struct {
		FromSource      bool   `json:"fromsource"`
		InstalledVer    string `json:"installedversion"`
		InstalledCommit string `json:"installedcommit"`
		CurrentVer      string `json:"currentversion"`
		CurrentCommit   string `json:"currentcommit"`
		UpdateAvailable bool   `json:"avail"`
	}
	var fromSource bool
	if installedVer == "" {
		fromSource = true
	}
	var uAvail bool
	if fromSource {
		uAvail = vars.CommitSHA != strings.TrimSpace(currentCommit)
	} else {
		uAvail = installedVer != strings.TrimSpace(currentVer)
	}
	verInfo := VersionInfo{
		FromSource:      fromSource,
		InstalledVer:    installedVer,
		InstalledCommit: vars.CommitSHA,
		CurrentVer:      strings.TrimSpace(currentVer),
		CurrentCommit:   strings.TrimSpace(currentCommit),
		UpdateAvailable: uAvail,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(verInfo)
}

func handleGenerateCerts(w http.ResponseWriter) {
	if err := botsetup.CreateCertCombo(); err != nil {
		http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, "done")
}

func saveCustomIntents() {
	customIntentJSONFile, _ := json.Marshal(vars.CustomIntents)
	os.WriteFile(vars.CustomIntentsPath, customIntentJSONFile, 0644)
}

func DisableCachingAndSniffing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate, max-age=0")
		w.Header().Set("pragma", "no-cache")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Expires", "0")
		next.ServeHTTP(w, r)
	})
}

func StartWebServer() {
	botsetup.RegisterSSHAPI()
	botsetup.RegisterBLEAPI()
	http.HandleFunc("/api/", apiHandler)
	http.HandleFunc("/session-certs/", certHandler)
	var webRoot http.Handler
	if runtime.GOOS == "darwin" && vars.Packaged {
		appPath, _ := os.Executable()
		webRoot = http.FileServer(http.Dir(filepath.Dir(appPath) + "/../Frameworks/chipper/webroot"))
	} else if runtime.GOOS == "android" || runtime.GOOS == "ios" {
		webRoot = http.FileServer(http.Dir(vars.AndroidPath + "/static/webroot"))
	} else {
		webRoot = http.FileServer(http.Dir("./webroot"))
	}
	http.Handle("/", DisableCachingAndSniffing(webRoot))
	fmt.Printf("Starting webserver at port " + vars.WebPort + " (http://localhost:" + vars.WebPort + ")\n")
	if err := http.ListenAndServe(":"+vars.WebPort, nil); err != nil {
		logger.Println("Error binding to " + vars.WebPort + ": " + err.Error())
		if vars.Packaged {
			logger.ErrMsg("FATAL: Wire-pod was unable to bind to port " + vars.WebPort + ". Another process is likely using it. Exiting.")
		}
		os.Exit(1)
	}
}

func GetLatestCommitSha() (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.github.com/repos/kercre123/wire-pod/commits", nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get commits: %s", resp.Status)
	}
	type Commit struct {
		Sha string `json:"sha"`
	}
	var commits []Commit
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return "", err
	}
	if len(commits) == 0 {
		return "", fmt.Errorf("no commits found")
	}
	return commits[0].Sha[:7], nil
}

func GetLatestReleaseTag(owner, repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	type Release struct {
		TagName string `json:"tag_name"`
	}
	var release Release
	if err := json.Unmarshal(body, &release); err != nil {
		return "", err
	}

	return release.TagName, nil
}

func certHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.Contains(r.URL.Path, "/session-certs/"):
		split := strings.Split(r.URL.Path, "/")
		if len(split) < 3 {
			http.Error(w, "must request a cert by esn (ex. /session-certs/00e20145)", http.StatusBadRequest)
			return
		}
		esn := split[2]
		fileBytes, err := os.ReadFile(path.Join(vars.SessionCertPath, esn))
		if err != nil {
			http.Error(w, "cert does not exist", http.StatusNotFound)
			return
		}
		w.Write(fileBytes)
	}
}

func anyEmpty(values ...string) bool {
	for _, v := range values {
		if v == "" {
			return true
		}
	}
	return false
}

func isValidLanguage(language string, validLanguages []string) bool {
	for _, lang := range validLanguages {
		if lang == language {
			return true
		}
	}
	return false
}

func isDownloadedLanguage(language string, downloadedLanguages []string) bool {
	for _, lang := range downloadedLanguages {
		if lang == language {
			return true
		}
	}
	return false
}

// handleTestEndpoint tests the connection to an API endpoint
func handleTestEndpoint(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Endpoint string `json:"endpoint"`
		ApiKey   string `json:"api_key"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	
	// Ensure the endpoint is valid
	endpoint := request.Endpoint
	if !strings.HasPrefix(endpoint, "http") {
		http.Error(w, "invalid endpoint URL", http.StatusBadRequest)
		return
	}
	
	// Create a test request to the health or models endpoint
	testEndpoint := endpoint
	if strings.HasSuffix(testEndpoint, "/v1") {
		testEndpoint += "/models"
	} else if !strings.HasSuffix(testEndpoint, "/models") {
		testEndpoint = strings.TrimSuffix(testEndpoint, "/") + "/v1/models"
	}
	
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", testEndpoint, nil)
	if err != nil {
		http.Error(w, "Error creating request: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Add API key if provided
	if request.ApiKey != "" {
		req.Header.Add("Authorization", "Bearer "+request.ApiKey)
	}
	
	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		// Connection error
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"message": "Connection failed: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()
	
	// Check response status
	statusCode := resp.StatusCode
	var result map[string]interface{}
	
	if statusCode >= 200 && statusCode < 300 {
		// Success - connection is working
		result = map[string]interface{}{
			"status": "success",
			"message": "Connection successful",
			"http_code": statusCode,
		}
	} else if statusCode == 401 || statusCode == 403 {
		// Authentication error
		result = map[string]interface{}{
			"status": "auth_error",
			"message": "Authentication failed: Invalid API key",
			"http_code": statusCode,
		}
	} else {
		// Other error
		bodyBytes, _ := io.ReadAll(resp.Body)
		result = map[string]interface{}{
			"status": "error",
			"message": "Endpoint error: " + resp.Status,
			"http_code": statusCode,
			"response": string(bodyBytes),
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
