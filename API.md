
# `github.com/xyproto/katago` Package API Documentation

The `katago` package provides an interface to interact with the KataGo analysis engine. This package allows you to send analysis requests to KataGo and receive detailed analysis responses.

## Overview

The `katago` package allows you to start a KataGo process, send analysis requests, and handle the responses. The main components of the package are:

- `AnalysisRequest`: Represents a request to analyze a position or a sequence of moves.
- `AnalysisResponse`: Represents the response from KataGo for an analysis request.
- `KataGo`: Represents a KataGo analysis engine instance.

## Installation

To install the `katago` package, use the following command:

```sh
go get github.com/xyproto/katago
```

## Usage

### Importing the Package

```go
import "github.com/xyproto/katago"
```

### Creating a KataGo Instance

To create a new KataGo instance, use the `NewKataGo` function. You need to provide the path to the configuration file and the model file.

```go
configFile := "path/to/analysis_example.cfg"
modelFile := "path/to/model.bin.gz"
katagoInstance, err := katago.NewKataGo(configFile, modelFile)
if err != nil {
    log.Fatalf("Failed to initialize KataGo: %v", err)
}
```

### Creating an Analysis Request

An `AnalysisRequest` specifies the details of the position or sequence of moves you want to analyze.

#### Fields of `AnalysisRequest`

- `ID` (string): An arbitrary string identifier for the query.
- `InitialStones` ([][2]string): Specifies stones already on the board at the start of the game. For example, these could be handicap stones.
- `Moves` ([][2]string): The moves that were played in the game, in the order they were played.
- `Rules` (string): Specify the rules for the game (e.g., "tromp-taylor").
- `Komi` (float64): The komi for the game.
- `BoardXSize` (int): The width of the board.
- `BoardYSize` (int): The height of the board.
- `MaxVisits` (int, optional): The maximum number of visits to use.
- `AnalyzeTurns` ([]int): Which turns of the game to analyze. 0 is the initial position, 1 is the position after `Moves[0]`, 2 is the position after `Moves[1]`, etc.

#### Example

```go
request := katago.AnalysisRequest{
    ID:            "example1",
    InitialStones: [][2]string{{"B", "Q16"}, {"W", "D4"}},
    Moves:         [][2]string{{"B", "D16"}},
    Rules:         "tromp-taylor",
    Komi:          7.5,
    BoardXSize:    19,
    BoardYSize:    19,
    MaxVisits:     1000,
    AnalyzeTurns:  []int{0, 1},
}
```

### Sending an Analysis Request

To send an analysis request, use the `Analyze` method of the `KataGo` instance. This method returns a slice of `AnalysisResponse`.

```go
responses, err := katagoInstance.Analyze([]katago.AnalysisRequest{request})
if err != nil {
    log.Fatalf("Failed to analyze request: %v", err)
}

for _, response := range responses {
    log.Printf("Received response: %v", response)
}
```

### Handling Analysis Responses

An `AnalysisResponse` contains the analysis results for the request. You can access various details, such as the move information and winrates.

```go
for _, response := range responses {
    log.Printf("Response ID: %s", response.ID)
    for _, moveInfo := range response.MoveInfos {
        log.Printf("Move: %s, Winrate: %f", moveInfo.Move, moveInfo.Winrate)
    }
}
```

### Closing the KataGo Instance

After you are done with the analysis, make sure to close the KataGo instance to release resources.

```go
if err := katagoInstance.Close(); err != nil {
    log.Fatalf("Failed to close KataGo: %v", err)
}
```

## Example

Here is a complete example demonstrating how to use the `katago` package:

```go
package main

import (
    "log"

    "github.com/xyproto/katago"
)

func main() {
    // Initialize KataGo instance
    configFile := "analysis_example.cfg"
    modelFile := "model.bin.gz"
    katagoInstance, err := katago.NewKataGo(configFile, modelFile)
    if err != nil {
        log.Fatalf("Failed to initialize KataGo: %v", err)
    }
    defer func() {
        if err := katagoInstance.Close(); err != nil {
            log.Fatalf("Failed to close KataGo: %v", err)
        }
    }()

    // Create an analysis request
    request := katago.AnalysisRequest{
        ID:            "example1",
        InitialStones: [][2]string{{"B", "Q16"}, {"W", "D4"}},
        Moves:         [][2]string{{"B", "D16"}},
        Rules:         "tromp-taylor",
        Komi:          7.5,
        BoardXSize:    19,
        BoardYSize:    19,
        MaxVisits:     1000,
        AnalyzeTurns:  []int{0, 1},
    }

    // Send the analysis request
    responses, err := katagoInstance.Analyze([]katago.AnalysisRequest{request})
    if err != nil {
        log.Fatalf("Failed to analyze request: %v", err)
    }

    // Handle the analysis responses
    for _, response := range responses {
        log.Printf("Response ID: %s", response.ID)
        for _, moveInfo := range response.MoveInfos {
            log.Printf("Move: %s, Winrate: %f", moveInfo.Move, moveInfo.Winrate)
        }
    }
}
```

## Analyzing Move 10 of a Given SGF File

Here's a simple example of how to analyze move 10 of a given SGF file.

### `main.go`

```go
package main

import (
    "log"
    "os"

    "github.com/xyproto/katago"
    "github.com/otrego/clamshell/sgf"
)

func main() {
    // Open SGF file
    sgfFile, err := os.Open("game.sgf")
    if err != nil {
        log.Fatalf("Failed to open SGF file: %v", err)
    }
    defer sgfFile.Close()

    // Parse SGF file
    tree, err := sgf.Parse(sgfFile)
    if err != nil {
        log.Fatalf("Failed to parse SGF file: %v", err)
    }

    // Extract moves from SGF file
    moves := extractMoves(tree.Root)

    // Initialize KataGo instance
    configFile := "analysis_example.cfg"
    modelFile := "model.bin.gz"
    katagoInstance, err := katago.NewKataGo(configFile, modelFile)
    if err != nil {
        log.Fatalf("Failed to initialize KataGo: %v", err)
    }
    defer func() {
        if err := katagoInstance.Close(); err != nil {
            log.Fatalf("Failed to close KataGo: %v", err)
        }
    }()

    // Create an analysis request for move 10
    request := katago.AnalysisRequest{
        ID:            "analyze_move_10",
        Moves:         moves[:10],
        Rules:         "tromp-taylor",
        Komi:          7.5,
        BoardXSize:    19,
        BoardYSize:    19,
        MaxVisits:     1000,
        AnalyzeTurns:  []int{10},
    }

    // Send the analysis request
    responses, err := katagoInstance.Analyze([]katago.AnalysisRequest{request})
    if err != nil {
        log.Fatalf("Failed to analyze request: %v", err)
    }

    // Handle the analysis responses
    for _, response := range responses {
        log.Printf("Response ID: %s", response.ID)
        for _, moveInfo := range response.MoveInfos {
            log.Printf("Move: %s, Winrate: %f", moveInfo.Move, moveInfo.Winrate)
        }
    }
}

// extractMoves extracts moves from an SGF node
func extractMoves(node *sgf.Node) [][2]string {
    var moves [][2]string
    for node != nil {
        for _, move := range node.Moves {
            moves = append(moves, [2]string{move.Color, move.Point.String()})
        }
        node = node.Next()
    }
    return moves
}
```

## API Reference

### `type AnalysisRequest`

```go
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
```

### `type AnalysisResponse`

```go
type AnalysisResponse struct {
    ID        string        `json:"id"`
    MoveInfos []MoveInfoExt `json:"moveInfos"`
}
```

### `type MoveInfoExt`

```go
type MoveInfoExt struct {
    Move    string  `json:"move"`
    Winrate float64 `json:"winrate"`
}
```

### `type KataGo`

```go
type KataGo struct {
    cmd    *exec.Cmd
    stdin  io.WriteCloser
    stdout *bufio.Reader
    stderr *bufio.Scanner
}
```

### `func NewKataGo(configFile, modelFile string) (*KataGo, error)`

```go
func NewKataGo(configFile, modelFile string) (*KataGo, error)
```

### `func (k *KataGo) Analyze(requests []AnalysisRequest) ([]AnalysisResponse, error)`

```go
func (k *KataGo) Analyze(requests []AnalysisRequest) ([]AnalysisResponse, error)
```

### `func (k *KataGo) Close() error`

```go
func (k *KataGo) Close() error
```
