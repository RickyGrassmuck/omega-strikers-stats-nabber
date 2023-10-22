package tracker

type GamePhase string

var ValidGamePhases = map[GamePhase]bool{
	"ArenaOverview":       true,
	"BanCelebration":      true,
	"BanSelect":           true,
	"CharacterPreSelect":  true,
	"CharacterSelect":     true,
	"FaceOffCountdown":    true,
	"FaceOffIntro":        true,
	"GoalCelebration":     true,
	"GoalScore":           true,
	"InGame":              true,
	"Intermission":        true,
	"IntermissionIntro":   true,
	"IntermissionMvp":     true,
	"IntermissionOutro":   true,
	"LoadoutSelect":       true,
	"None":                true,
	"PostGameCelebration": true,
	"PostGameSummary":     true,
	"PreGame":             true,
	"VersusScreen":        true,
}

func IsValidGamePhase(phase string) bool {
	_, exists := ValidGamePhases[GamePhase(phase)]
	return exists
}
