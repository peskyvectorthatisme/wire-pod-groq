package wirepod_whisper

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	sr "github.com/kercre123/wire-pod/chipper/pkg/wirepod/speechrequest"
	"github.com/orcaman/writerseeker"
)

var Name string = "whisper"

type openAiResp struct {
	Text string `json:"text"`
}

// Provider defines which API provider to use
type Provider struct {
	Name     string
	BaseURL  string
	Key      string
}

var currentProvider Provider

func Init() error {
	// Check if Groq config file exists and load it
	if _, err := os.Stat("groq_config.json"); err == nil {
		configData, err := os.ReadFile("groq_config.json")
		if err == nil {
			var config struct {
				Knowledge struct {
					Provider string `json:"provider"`
					Key      string `json:"key"`
					Endpoint string `json:"endpoint"`
				} `json:"knowledge"`
			}
			if json.Unmarshal(configData, &config) == nil && config.Knowledge.Provider == "groq" {
				currentProvider = Provider{
					Name:     "groq",
					BaseURL:  config.Knowledge.Endpoint,
					Key:      config.Knowledge.Key,
				}
				logger.Println("Using Groq for Whisper API (from config file)")
				
				// Update the STT configuration to match
				vars.APIConfig.STT.Provider = "groq"
				vars.APIConfig.STT.APIKey = config.Knowledge.Key
				vars.APIConfig.STT.Endpoint = config.Knowledge.Endpoint
				vars.WriteConfigToDisk()
				
				// Remove the temporary config file
				os.Remove("groq_config.json")
				return nil
			}
		}
	}

	// First check STT configuration
	if vars.APIConfig.STT.Provider != "" && vars.APIConfig.STT.APIKey != "" {
		baseURL := vars.APIConfig.STT.Endpoint
		if baseURL == "" {
			if vars.APIConfig.STT.Provider == "groq" {
				baseURL = "https://api.groq.com/openai/v1"
			} else if vars.APIConfig.STT.Provider == "openai" {
				baseURL = "https://api.openai.com/v1"
			} else {
				// If it's a custom provider, we must have an endpoint
				return fmt.Errorf("custom provider requires an endpoint URL")
			}
		}
		
		currentProvider = Provider{
			Name:     vars.APIConfig.STT.Provider,
			BaseURL:  baseURL,
			Key:      vars.APIConfig.STT.APIKey,
		}
		logger.Println(fmt.Sprintf("Using %s for Whisper API (from STT config)", vars.APIConfig.STT.Provider))
		return nil
	}
	
	// Check if provider is set in the Knowledge configuration (legacy support)
	if vars.APIConfig.Knowledge.Provider == "groq" && vars.APIConfig.Knowledge.Key != "" && vars.APIConfig.Knowledge.Endpoint != "" {
		currentProvider = Provider{
			Name:     "groq",
			BaseURL:  vars.APIConfig.Knowledge.Endpoint,
			Key:      vars.APIConfig.Knowledge.Key,
		}
		
		// Update the STT configuration to match
		vars.APIConfig.STT.Provider = "groq"
		vars.APIConfig.STT.APIKey = vars.APIConfig.Knowledge.Key
		vars.APIConfig.STT.Endpoint = vars.APIConfig.Knowledge.Endpoint
		vars.WriteConfigToDisk()
		
		logger.Println("Using Groq for Whisper API (migrated from Knowledge config)")
	} else if os.Getenv("OPENAI_KEY") != "" {
		// Legacy support for environment variable
		currentProvider = Provider{
			Name:     "openai",
			BaseURL:  "https://api.openai.com/v1",
			Key:      os.Getenv("OPENAI_KEY"),
		}
		
		// Update the STT configuration to match
		vars.APIConfig.STT.Provider = "openai"
		vars.APIConfig.STT.APIKey = os.Getenv("OPENAI_KEY")
		vars.APIConfig.STT.Endpoint = "https://api.openai.com/v1"
		vars.WriteConfigToDisk()
		
		logger.Println("Using OpenAI for Whisper API (migrated from environment)")
	} else if vars.APIConfig.Knowledge.Provider == "openai" && vars.APIConfig.Knowledge.Key != "" {
		currentProvider = Provider{
			Name:     "openai",
			BaseURL:  "https://api.openai.com/v1",
			Key:      vars.APIConfig.Knowledge.Key,
		}
		
		// Update the STT configuration to match
		vars.APIConfig.STT.Provider = "openai"
		vars.APIConfig.STT.APIKey = vars.APIConfig.Knowledge.Key
		vars.APIConfig.STT.Endpoint = "https://api.openai.com/v1"
		vars.WriteConfigToDisk()
		
		logger.Println("Using OpenAI for Whisper API (migrated from Knowledge config)")
	} else {
		logger.Println("No valid Whisper API configuration found. Please configure a provider in the STT settings.")
		return fmt.Errorf("no valid API provider configured for Whisper")
	}
	return nil
}

func pcm2wav(in io.Reader) []byte {

	// Output file.
	out := &writerseeker.WriterSeeker{}

	// 8 kHz, 16 bit, 1 channel, WAV.
	e := wav.NewEncoder(out, 16000, 16, 1, 1)

	// Create new audio.IntBuffer.
	audioBuf, err := newAudioIntBuffer(in)
	if err != nil {
		logger.Println(err)
	}
	// Write buffer to output file. This writes a RIFF header and the PCM chunks from the audio.IntBuffer.
	if err := e.Write(audioBuf); err != nil {
		logger.Println(err)
	}
	if err := e.Close(); err != nil {
		logger.Println(err)
	}
	outBuf := new(bytes.Buffer)
	io.Copy(outBuf, out.BytesReader())
	return outBuf.Bytes()
}

func newAudioIntBuffer(r io.Reader) (*audio.IntBuffer, error) {
	buf := audio.IntBuffer{
		Format: &audio.Format{
			NumChannels: 1,
			SampleRate:  16000,
		},
	}
	for {
		var sample int16
		err := binary.Read(r, binary.LittleEndian, &sample)
		switch {
		case err == io.EOF:
			return &buf, nil
		case err != nil:
			return nil, err
		}
		buf.Data = append(buf.Data, int(sample))
	}
}

func makeAPIRequest(in []byte) (string, error) {
	if currentProvider.Name == "" {
		return "", fmt.Errorf("no API provider configured")
	}

	// Construct the full endpoint URL
	endpointURL := currentProvider.BaseURL
	if !strings.HasSuffix(endpointURL, "/") {
		endpointURL += "/"
	}
	if !strings.HasSuffix(endpointURL, "audio/transcriptions") {
		endpointURL = strings.TrimSuffix(endpointURL, "/") + "/audio/transcriptions"
	}

	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	w.WriteField("model", "whisper-1")
	sendFile, _ := w.CreateFormFile("file", "audio.wav")
	sendFile.Write(in)
	w.Close()

	httpReq, _ := http.NewRequest("POST", endpointURL, buf)
	httpReq.Header.Set("Content-Type", w.FormDataContentType())
	httpReq.Header.Set("Authorization", "Bearer "+currentProvider.Key)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		logger.Println("API request error:", err)
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		logger.Println("API error response:", resp.Status, string(responseBody))
		return "", fmt.Errorf("API error: %s", resp.Status)
	}

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var aiResponse openAiResp
	if err := json.Unmarshal(response, &aiResponse); err != nil {
		logger.Println("Error unmarshaling response:", err)
		return "", err
	}

	return aiResponse.Text, nil
}

func STT(req sr.SpeechRequest) (string, error) {
	logger.Println("(Bot " + req.Device + ", Whisper) Processing...")
	speechIsDone := false
	var err error
	for {
		_, err = req.GetNextStreamChunk()
		if err != nil {
			return "", err
		}
		// has to be split into 320 []byte chunks for VAD
		speechIsDone, _ = req.DetectEndOfSpeech()
		if speechIsDone {
			break
		}
	}

	pcmBufTo := &writerseeker.WriterSeeker{}
	pcmBufTo.Write(req.DecodedMicData)
	pcmBuf := pcm2wav(pcmBufTo.BytesReader())

	transcribedText, err := makeAPIRequest(pcmBuf)
	if err != nil {
		logger.Println("Error from API:", err)
		return "", err
	}
	
	transcribedText = strings.ToLower(transcribedText)
	logger.Println("Bot " + req.Device + " Transcribed text: " + transcribedText)
	return transcribedText, nil
}
