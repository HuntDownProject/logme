package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/projectdiscovery/gologger"
)

func main() {
	// Definindo os argumentos de linha de comando
	server := flag.String("server", "", "Splunk server address (ex: https://splunk.example.com:8088/services/collector/event)")
	token := flag.String("token", "", "Access token")
	sourcetype := flag.String("sourcetype", "_json", "Sourcetype of the event")
	file := flag.String("file", "", "File path to be sent to Splunk HEC")
	flag.Parse()

	if *server == "" || *token == "" || *file == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	fileContent, err := os.ReadFile(*file)
	if err != nil {
		gologger.Fatal().Msgf("Failed while read the file: %v\n", err)
	}

	if strings.HasSuffix(*server, "/services/collector/event") {
		payload := map[string]interface{}{
			"event":      string(fileContent),
			"sourcetype": sourcetype,
		}
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			gologger.Fatal().Msgf("Failed to create payload: %v\n", err)
		}
		fileContent = jsonPayload
	}

	// Enviando o arquivo para o Splunk
	url := *server
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(fileContent))
	if err != nil {
		gologger.Fatal().Msgf("Failed to create request: %s\n", err)
	}

	// Adicionando os headers necessários
	req.Header.Set("Authorization", "Splunk "+*token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		gologger.Fatal().Msgf("Failed to send request: %s\n", err)
	}
	defer resp.Body.Close()

	// Lendo a resposta
	body, _ := io.ReadAll(resp.Body)
	gologger.Info().Msgf("Server response: %s\n", string(body))
}
