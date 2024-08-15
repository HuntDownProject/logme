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
	break_line := flag.Bool("bl", false, "Break line in the file content")
	flag.Parse()

	if *server == "" || *token == "" || *file == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	fileContent, err := os.ReadFile(*file)
	if err != nil {
		gologger.Fatal().Msgf("Failed while read the file: %v\n", err)
	}

	var lines []string
	client := &http.Client{}

	if break_line == nil || *break_line {
		lines = strings.Split(string(fileContent), "\n")

		for _, line := range lines {
			var buffer bytes.Buffer
			buffer.WriteString(line + "\n")
			var bufferBytes = buffer.Bytes()
			if len(bufferBytes) > 1 {
				sendToSplunk(client, server, token, sourcetype, buffer.Bytes())
			}
		}
	} else {
		sendToSplunk(client, server, token, sourcetype, fileContent)
	}

}

func createPayload(server *string, fileContent []byte, sourcetype *string) []byte {
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

	return fileContent
}

func sendToSplunk(client *http.Client, server *string, token *string, sourcetype *string, fileContent []byte) {

	fileContent = createPayload(server, fileContent, sourcetype)

	req, err := http.NewRequest("POST", *server, bytes.NewBuffer(fileContent))
	if err != nil {
		gologger.Fatal().Msgf("Failed to create request: %s\n", err)
	}

	req.Header.Set("Authorization", "Splunk "+*token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		gologger.Fatal().Msgf("Failed to send request: %s\n", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	gologger.Info().Msgf("Server response: %s\n", string(body))
}
