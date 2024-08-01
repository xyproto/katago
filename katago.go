package katago

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
)

// AnalysisRequest represents a request to analyze a position or a sequence of moves
type AnalysisRequest struct {
	ID            string      `json:"id"`
	InitialStones [][2]string `json:"initialStones,omitempty"`
	Moves         [][2]string `json:"moves"`
	Rules         string      `json:"rules"`
	Komi          float64     `json:"komi"`
	BoardXSize    int         `json:"boardXSize"`
	BoardYSize    int         `json:"boardYSize"`
	MaxVisits     int         `json:"maxVisits,omitempty"`
	AnalyzeTurns  []int       `json:"analyzeTurns"`
}

// AnalysisResponse represents the response from KataGo for an analysis request
type AnalysisResponse struct {
	ID        string        `json:"id"`
	MoveInfos []MoveInfoExt `json:"moveInfos"`
}

// MoveInfoExt represents the extended information about a move analyzed by KataGo
type MoveInfoExt struct {
	Move    string  `json:"move"`
	Winrate float64 `json:"winrate"`
}

// KataGo represents a KataGo analysis engine instance
type KataGo struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
	stderr *bufio.Scanner
}

// NewKataGo creates a new KataGo analysis engine instance
func NewKataGo(configFile, modelFile string) (*KataGo, error) {
	cmd := exec.Command("katago", "analysis", "-config", configFile, "-model", modelFile)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stderr: %v", err)
	}

	k := &KataGo{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewReader(stdout),
		stderr: bufio.NewScanner(stderr),
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start KataGo: %v", err)
	}

	go k.readStderr()

	return k, nil
}

// readStderr reads from KataGo's stderr for logging purposes
func (k *KataGo) readStderr() {
	for k.stderr.Scan() {
		fmt.Printf("KataGo stderr: %s\n", k.stderr.Text())
	}
	if err := k.stderr.Err(); err != nil {
		fmt.Printf("Error reading stderr: %v\n", err)
	}
}

// Analyze sends multiple analysis requests to KataGo and returns the responses
func (k *KataGo) Analyze(requests []AnalysisRequest) ([]AnalysisResponse, error) {
	var responses []AnalysisResponse
	responseMap := make(map[string]AnalysisResponse)

	for _, request := range requests {
		// Log the request being sent
		log.Printf("Sending request: %v", request)

		// Send analysis request to KataGo
		requestJSON, err := json.Marshal(request)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %v", err)
		}
		fmt.Fprintf(k.stdin, "%s\n", requestJSON)
	}

	for len(responseMap) < len(requests) {
		// Read response from KataGo
		responseJSON, err := k.stdout.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("error reading response: %v", err)
		}

		var response AnalysisResponse
		if err := json.Unmarshal([]byte(responseJSON), &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %v", err)
		}

		// Log the response received
		log.Printf("Received response: %v", response)
		responseMap[response.ID] = response
	}

	for _, request := range requests {
		responses = append(responses, responseMap[request.ID])
	}

	return responses, nil
}

// Close shuts down the KataGo process by closing its stdin
func (k *KataGo) Close() error {
	if err := k.stdin.Close(); err != nil {
		return fmt.Errorf("failed to close KataGo stdin: %v", err)
	}
	return k.cmd.Wait() // Wait for the process to exit cleanly
}
