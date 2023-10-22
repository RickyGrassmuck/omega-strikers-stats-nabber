package main

import (
	"fmt"
	"image/png"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/RickyGrassmuck/omega-strikers-stat-nabber/tracker"
	"github.com/kbinani/screenshot"
)

func captureScreen() {
	n := screenshot.NumActiveDisplays()

	for i := 0; i < n; i++ {
		bounds := screenshot.GetDisplayBounds(i)

		img, err := screenshot.CaptureRect(bounds)
		if err != nil {
			panic(err)
		}
		fileName := fmt.Sprintf("data/%d_%dx%d.png", i, bounds.Dx(), bounds.Dy())
		file, _ := os.Create(fileName)
		defer file.Close()
		png.Encode(file, img)

		fmt.Printf("#%d : %v \"%s\"\n", i, bounds, fileName)
	}
}

func main() {
	logDir := "C:/Users/TheBoss/AppData/Local/OmegaStrikers/Saved/Logs"
	logName := "OmegaStrikers.log"
	if len(os.Args) > 1 {
		logName = os.Args[1]
	}
	logFile := path.Join(logDir, logName)

	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		fmt.Printf("Log file %s does not exist\n", logFile)
		os.Exit(1)
	}
	log.Printf("[DEBUG] - Log File Found: %s", logFile)

	logLinesReceiver := make(chan tracker.LogLine)
	defer close(logLinesReceiver)

	go tracker.TailLogLines(logFile, logLinesReceiver)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case logLine := <-logLinesReceiver:
			if logLine.ParsedData != nil {
				switch logLine.MsgType {
				case tracker.MsgTypeName("MatchmakingStatus"):
					// log.Printf("Found %s - Format: %s - data: %#+v\n", logLine.MsgType, logLine.ParsedDataType, logLine.ParsedData)
				case tracker.MsgTypeName("GamePhaseChange"):
					// log.Printf("Found %s - Format: %s - data: %#+v\n", logLine.MsgType, logLine.ParsedDataType, logLine.ParsedData)
				case tracker.MsgTypeName("GoalScored"):
					log.Printf("%s - data: %#+v\n", logLine.ParsedDataType, logLine.ParsedData)
				case tracker.MsgTypeName("SetResult"):
					log.Printf("%s - data: %#+v\n", logLine.ParsedDataType, logLine.ParsedData)
				case tracker.MsgTypeName("MatchResult"):
					log.Printf("%s - data: %#+v\n", logLine.ParsedDataType, logLine.ParsedData)
				}
			}

		case <-sigChan:
			log.Printf("[DEBUG] - Received Signal, Exiting")
			return
		}
	}
}
