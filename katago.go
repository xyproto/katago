package katago

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
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
	cmd        *exec.Cmd
	stdin      io.Writer
	stdout     *bufio.Reader
	stderr     *bufio.Scanner
	requestCh  chan AnalysisRequest
	responseCh chan AnalysisResponse
	responses  map[string]chan AnalysisResponse
	mu         sync.Mutex
	wg         sync.WaitGroup
	closeCh    chan struct{}
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
		cmd:        cmd,
		stdin:      stdin,
		stdout:     bufio.NewReader(stdout),
		stderr:     bufio.NewScanner(stderr),
		requestCh:  make(chan AnalysisRequest),
		responseCh: make(chan AnalysisResponse),
		responses:  make(map[string]chan AnalysisResponse),
		closeCh:    make(chan struct{}),
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start KataGo: %v", err)
	}

	k.wg.Add(1)
	go k.run()

	return k, nil
}

// run handles the communication with the KataGo process
func (k *KataGo) run() {
	defer k.wg.Done()

	// Read stderr to debug any issues
	go func() {
		for k.stderr.Scan() {
			fmt.Printf("KataGo stderr: %s\n", k.stderr.Text())
		}
		if err := k.stderr.Err(); err != nil {
			fmt.Printf("Error reading stderr: %v\n", err)
		}
	}()

	for {
		select {
		case request := <-k.requestCh:
			// Log the request being sent
			log.Printf("Sending request: %v", request)

			// Send analysis request to KataGo
			requestJSON, err := json.Marshal(request)
			if err != nil {
				log.Fatalf("failed to marshal request: %v", err)
			}
			fmt.Fprintf(k.stdin, "%s\n", requestJSON)

			// Read response from KataGo
			responseJSON, err := k.stdout.ReadString('\n')
			if err != nil {
				log.Fatalf("error reading response: %v", err)
			}

			var response AnalysisResponse
			if err := json.Unmarshal([]byte(responseJSON), &response); err != nil {
				log.Fatalf("failed to unmarshal response: %v", err)
			}

			// Log the response received
			log.Printf("Received response: %v", response)

			// Send response to the correct channel
			k.mu.Lock()
			if ch, ok := k.responses[response.ID]; ok {
				ch <- response
				close(ch) // Signal that no more data will be sent
				delete(k.responses, response.ID)
			}
			k.mu.Unlock()

		case <-k.closeCh:
			return
		}
	}
}

// Analyze sends an analysis request to KataGo and returns the response
func (k *KataGo) Analyze(request AnalysisRequest) (AnalysisResponse, error) {
	responseCh := make(chan AnalysisResponse)
	k.mu.Lock()
	k.responses[request.ID] = responseCh
	k.mu.Unlock()
	k.requestCh <- request
	response := <-responseCh
	return response, nil
}

// Close shuts down the KataGo process
func (k *KataGo) Close() error {
	close(k.closeCh)
	k.wg.Wait()
	if err := k.cmd.Process.Kill(); err != nil {
		return fmt.Errorf("failed to kill KataGo process: %v", err)
	}
	return nil
}
