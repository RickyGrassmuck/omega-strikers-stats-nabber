package tracker

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type MsgTypeName string
type MatchmakingStatus map[string]interface{}
type GoalScoredEvent struct {
	Team     string
	NewScore int
}

type MsgType struct {
	Regex          string
	Proccesor      func(match []string, logline *LogLine) (success bool, err error)
	ParsedDataType reflect.Type
}

var MessageTypes map[MsgTypeName]MsgType = map[MsgTypeName]MsgType{
	"Undefined": {
		Regex:          "",
		Proccesor:      processUndefined,
		ParsedDataType: reflect.TypeOf(reflect.String),
	},
	"MatchmakingStatus": {
		Regex:          `(?m).*Matchmaking Status: (.*)`,
		Proccesor:      processMatchmakingStatus,
		ParsedDataType: reflect.TypeOf(MatchmakingStatus{}),
	},

	"GamePhaseChange": {
		Regex:          `(?m).*APMGameState::PerformCurrentMatchPhaseEvents.*Current\[(.*)\]`,
		Proccesor:      processGamePhaseChange,
		ParsedDataType: reflect.TypeOf(GamePhase("")),
	},

	"GoalScored": {
		Regex:          `(?m)([a-zA-Z]+)'s NumPointsThisSet changed from \d to (\d)`,
		Proccesor:      processGoalScored,
		ParsedDataType: reflect.TypeOf(GoalScoredEvent{}),
	},

	"SetResult": {
		Regex:          `(?m)TeamThatWonSet changed from '<unset>' to 'EAssignedTeam::(.*)'`,
		Proccesor:      processSetResult,
		ParsedDataType: reflect.TypeOf(SetResult{}),
	},

	"MatchResult": {
		Regex:          `(?m)TeamThatWonMatch changed from '<unset>' to 'EAssignedTeam::(.*)'`,
		Proccesor:      processMatchResult,
		ParsedDataType: reflect.TypeOf(MatchResults{}),
	},
}

func (m MsgTypeName) GetParsedDataType() reflect.Type {
	return MessageTypes[m].ParsedDataType
}

var LogLineRegexp = regexp.MustCompile(`(?m)^\[(.*)\]\[.*\]([a-zA-Z0-9]+): (.*)`)

var IgnoredMessagePatterns = []string{
	`(?m)^Warning:.*`,
	`(?m)^Error:.*`,
	`(?m)^Shutting down.*`,
}

// TailLogLines accepts a file path and returns a channel of LogLine structs
// representing the lines in the file.  The channel is closed when the end of
// the file is reached.
func TailLogLines(filePath string, outChan chan LogLine) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		parsedLine, ok := parseLogLine(line)
		if !ok {
			continue
		}
		outChan <- parsedLine
	}
}

func parseLogLine(line string) (LogLine, bool) {
	var parsedLine LogLine

	matches := LogLineRegexp.FindStringSubmatch(line)
	if len(matches) != 4 {
		return LogLine{}, false
	}
	parsedLine = LogLine{
		Timestamp:      matches[1],
		Module:         matches[2],
		RawMessage:     matches[3],
		MsgType:        MsgTypeName("Undefined"),
		ParsedDataType: MsgTypeName("Undefined").GetParsedDataType(),
	}

	if parsedLine.isIgnored() {
		return parsedLine, false
	}
	parsedLine.setMsgType()
	return parsedLine, true
}

func (l *LogLine) isIgnored() bool {
	for _, pattern := range IgnoredMessagePatterns {
		match, _ := regexp.MatchString(pattern, l.RawMessage)
		if match {
			return true
		}
	}
	return false
}

func (l *LogLine) setMsgType() {
	for msgType, handler := range MessageTypes {
		if msgType == "Undefined" {
			continue
		}

		re := regexp.MustCompile(handler.Regex)
		match := re.FindStringSubmatch(l.RawMessage)
		if match != nil {
			l.MsgType = msgType
			l.ParsedDataType = handler.ParsedDataType
			handler.Proccesor(match, l)
			return
		}
	}
	MessageTypes["Undefined"].Proccesor([]string{l.RawMessage}, l)
}

// processUndefined is the default message processor.  It simply returns the
// raw message as the parsed data.
func processUndefined(data []string, logLine *LogLine) (success bool, err error) {
	logLine.ParsedData = strings.Join(data, " ")
	return true, nil
}

func processSetResult(data []string, logLine *LogLine) (success bool, err error) {
	winner := data[1]
	event := SetResult{
		Winner: Teams[winner],
	}

	logLine.ParsedData = event
	return true, nil
}

func processMatchResult(data []string, logLine *LogLine) (success bool, err error) {
	winner := data[1]
	event := SetResult{
		Winner: Teams[winner],
	}

	logLine.ParsedData = event
	return true, nil
}

func processGoalScored(data []string, logLine *LogLine) (success bool, err error) {
	team, scoreStr := data[1], data[2]
	score, err := strconv.Atoi(scoreStr)
	if err != nil {
		return false, err
	}
	if score == 0 {
		return false, fmt.Errorf("Score reset to 0, ignoring")
	} else if score > 3 || score < 0 {
		return false, fmt.Errorf("Invalid score detected: %d", score)
	} else {
		event := GoalScoredEvent{
			Team:     Teams[team],
			NewScore: score,
		}

		logLine.ParsedData = event
		return true, nil
	}
}

// processMatchmakingStatus extracts the matchmaking status json string from
// the raw message and parses it into the LogLine.ParsedData field as a map[string]interface{}.
func processMatchmakingStatus(data []string, logLine *LogLine) (success bool, err error) {
	mmStatusRaw := data[1]
	var mmStatus map[string]interface{}
	err = json.Unmarshal([]byte(mmStatusRaw), &mmStatus)
	if err != nil {
		return false, err
	}
	logLine.ParsedData = mmStatus
	return true, err
}

// processGamePhaseChange extracts the current game phase from the raw message
// and stores it in the LogLine.ParsedData field as a GamePhase.
func processGamePhaseChange(data []string, logLine *LogLine) (success bool, err error) {
	curPhaseName, found := strings.CutPrefix(data[1], "EMatchPhase::")
	if !found {
		return false, fmt.Errorf("Current phase not found in: %s", data)
	}
	if !IsValidGamePhase(curPhaseName) {
		return false, fmt.Errorf("Unknown Game Phase Detected: %s", curPhaseName)
	}
	logLine.ParsedData = GamePhase(curPhaseName)
	return
}
