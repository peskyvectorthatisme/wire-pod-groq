package processreqs

import (
	"fmt"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	sr "github.com/kercre123/wire-pod/chipper/pkg/wirepod/speechrequest"
	ttr "github.com/kercre123/wire-pod/chipper/pkg/wirepod/ttr"
)

// Server stores the config
type Server struct{}

var VoiceProcessor = ""

type JsonIntent struct {
	Name              string   `json:"name"`
	Keyphrases        []string `json:"keyphrases"`
	RequireExactMatch bool     `json:"requiresexact"`
}

var sttLanguage string = "en-US"

// speech-to-text
var sttHandler func(sr.SpeechRequest) (string, error)

// speech-to-intent (rhino)
var stiHandler func(sr.SpeechRequest) (string, map[string]string, error)

var isSti bool = false

func ReloadVosk() {
	// Initialize the selected STT service based on the current configuration
	if vars.APIConfig.STT.Service == "vosk" || vars.APIConfig.STT.Service == "whisper.cpp" || vars.APIConfig.STT.Service == "whisper" || vars.APIConfig.STT.Service == "coqui" || vars.APIConfig.STT.Service == "leopard" {
		// Restart the STT service with the new configuration
		logger.Println("Restarting STT service: " + vars.APIConfig.STT.Service)
		vars.SttInitFunc()
		vars.IntentList, _ = vars.LoadIntents()
	} else {
		logger.Println("Unknown STT service: " + vars.APIConfig.STT.Service)
	}
}

// New returns a new server
func New(InitFunc func() error, SttHandler interface{}, voiceProcessor string) (*Server, error) {

	// Decide the TTS language
	if voiceProcessor != "vosk" && voiceProcessor != "whisper.cpp" {
		vars.APIConfig.STT.Language = "en-US"
	}
	sttLanguage = vars.APIConfig.STT.Language
	vars.IntentList, _ = vars.LoadIntents()
	logger.Println("Initiating " + voiceProcessor + " voice processor with language " + sttLanguage)
	vars.SttInitFunc = InitFunc
	err := InitFunc()
	if err != nil {
		return nil, err
	}

	// SttHandler can either be `func(sr.SpeechRequest) (string, error)` or `func (sr.SpeechRequest) (string, map[string]string, error)`
	// second one exists to accomodate Rhino

	// check function type
	if str, is := SttHandler.(func(sr.SpeechRequest) (string, error)); is {
		sttHandler = str
	} else if str, is := SttHandler.(func(sr.SpeechRequest) (string, map[string]string, error)); is {
		stiHandler = str
		isSti = true
	} else {
		return nil, fmt.Errorf("stthandler not of correct type")
	}

	// Initiating the chosen voice processor and load intents from json
	VoiceProcessor = voiceProcessor

	// Load plugins
	ttr.LoadPlugins()

	return &Server{}, err
}
