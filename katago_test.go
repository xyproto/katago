package katago

import (
	"log"
	"os"
	"sync"
	"testing"
)

var (
	katagoInstance *KataGo
	initOnce       sync.Once
)

func initKataGo(t *testing.T) *KataGo {
	t.Helper()

	initOnce.Do(func() {
		var err error
		configFile := "analysis_example.cfg"
		modelFile := "model.bin.gz"
		katagoInstance, err = NewKataGo(configFile, modelFile)
		if err != nil {
			t.Fatalf("Failed to initialize KataGo: %v", err)
		}
	})

	return katagoInstance
}

func cleanupKataGo(t *testing.T) {
	t.Helper()

	if katagoInstance != nil {
		if err := katagoInstance.Close(); err != nil {
			t.Fatalf("Failed to close KataGo: %v", err)
		}
	}
}

func TestMain(m *testing.M) {
	// Setup code
	katagoInstance = initKataGo(&testing.T{})
	code := m.Run()
	// Cleanup code
	cleanupKataGo(&testing.T{})
	// Exit with the proper code
	os.Exit(code)
}

func TestKataGoAnalyze(t *testing.T) {
	katago := initKataGo(t)
	defer cleanupKataGo(t)

	// Create an analysis request
	request := AnalysisRequest{
		ID:            "test1",
		InitialStones: [][2]string{{"B", "Q16"}},
		Moves:         [][2]string{{"W", "D4"}},
		Rules:         "tromp-taylor",
		Komi:          7.5,
		BoardXSize:    19,
		BoardYSize:    19,
		AnalyzeTurns:  []int{0, 1},
	}

	// Send the analysis request
	log.Println("Sending analysis request for TestKataGoAnalyze")
	response, err := katago.Analyze(request)
	if err != nil {
		t.Fatalf("Failed to analyze request: %v", err)
	}

	// Validate the response
	log.Printf("Received response for TestKataGoAnalyze: %v", response)
	if response.ID != request.ID {
		t.Errorf("Expected response ID %s, got %s", request.ID, response.ID)
	}
	if len(response.MoveInfos) == 0 {
		t.Errorf("Expected move infos in response, got none")
	}
	for _, moveInfo := range response.MoveInfos {
		if moveInfo.Move == "" {
			t.Errorf("Expected move info to have a move, got empty")
		}
		if moveInfo.Winrate < 0 || moveInfo.Winrate > 1 {
			t.Errorf("Expected winrate between 0 and 1, got %f", moveInfo.Winrate)
		}
	}
}

func TestKataGoMultipleRequests(t *testing.T) {
	katago := initKataGo(t)
	defer cleanupKataGo(t)

	// Create multiple analysis requests
	requests := []AnalysisRequest{
		{
			ID:            "test1",
			InitialStones: [][2]string{{"B", "Q16"}},
			Moves:         [][2]string{{"W", "D4"}},
			Rules:         "tromp-taylor",
			Komi:          7.5,
			BoardXSize:    19,
			BoardYSize:    19,
			AnalyzeTurns:  []int{0, 1},
		},
		{
			ID:            "test2",
			InitialStones: [][2]string{{"B", "Q4"}},
			Moves:         [][2]string{{"W", "D16"}},
			Rules:         "tromp-taylor",
			Komi:          7.5,
			BoardXSize:    19,
			BoardYSize:    19,
			AnalyzeTurns:  []int{0, 1},
		},
	}

	// Sequentially send the analysis requests and store responses
	responses := make(map[string]AnalysisResponse)
	for _, request := range requests {
		log.Printf("Sending analysis request: %v", request)
		response, err := katago.Analyze(request)
		if err != nil {
			t.Fatalf("Failed to analyze request %s: %v", request.ID, err)
		}
		responses[response.ID] = response
		log.Printf("Received response for request %s: %v", request.ID, response)
	}

	// Validate responses
	for _, request := range requests {
		response, exists := responses[request.ID]
		if !exists {
			t.Fatalf("No response found for request ID %s", request.ID)
		}
		if response.ID != request.ID {
			t.Errorf("Expected response ID %s, got %s", request.ID, response.ID)
		}
		if len(response.MoveInfos) == 0 {
			t.Errorf("Expected move infos in response for request %s, got none", request.ID)
		}
		for _, moveInfo := range response.MoveInfos {
			if moveInfo.Move == "" {
				t.Errorf("Expected move info to have a move for request %s, got empty", request.ID)
			}
			if moveInfo.Winrate < 0 || moveInfo.Winrate > 1 {
				t.Errorf("Expected winrate between 0 and 1 for request %s, got %f", request.ID, moveInfo.Winrate)
			}
		}
	}
}
