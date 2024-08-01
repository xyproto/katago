package katago

import (
	"log"
	"testing"
)

func initKataGo(t *testing.T) *KataGo {
	t.Helper()

	var err error
	configFile := "analysis_example.cfg"
	modelFile := "model.bin.gz"
	katagoInstance, err := NewKataGo(configFile, modelFile)
	if err != nil {
		t.Fatalf("Failed to initialize KataGo: %v", err)
	}

	return katagoInstance
}

func cleanupKataGo(t *testing.T, k *KataGo) {
	t.Helper()

	if k != nil {
		if err := k.Close(); err != nil {
			t.Fatalf("Failed to close KataGo: %v", err)
		}
	}
}

func TestKataGoAnalyze(t *testing.T) {
	katago := initKataGo(t)
	defer cleanupKataGo(t, katago)

	// Create an analysis request
	requests := []AnalysisRequest{
		{
			ID:            "test1",
			InitialStones: [][2]string{{"B", "Q16"}},
			Moves:         [][2]string{{"W", "D4"}},
			Rules:         "tromp-taylor",
			Komi:          7.5,
			BoardXSize:    19,
			BoardYSize:    19,
			MaxVisits:     1000,
			AnalyzeTurns:  []int{0, 1},
		},
	}

	// Send the analysis request
	log.Println("Sending analysis request for TestKataGoAnalyze")
	responses, err := katago.Analyze(requests)
	if err != nil {
		t.Fatalf("Failed to analyze request: %v", err)
	}

	// Validate the response
	response := responses[0]
	log.Printf("Received response for TestKataGoAnalyze: %v", response)
	if response.ID != requests[0].ID {
		t.Errorf("Expected response ID %s, got %s", requests[0].ID, response.ID)
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
	defer cleanupKataGo(t, katago)

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
			MaxVisits:     1000,
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
			MaxVisits:     1000,
			AnalyzeTurns:  []int{0, 1},
		},
	}

	// Send the analysis requests
	log.Println("Sending multiple analysis requests for TestKataGoMultipleRequests")
	responses, err := katago.Analyze(requests)
	if err != nil {
		t.Fatalf("Failed to analyze requests: %v", err)
	}

	// Validate responses
	for i, request := range requests {
		response := responses[i]
		log.Printf("Received response for request %s: %v", request.ID, response)
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
