package stringpatterns

import "github.com/AngadVM/goprofiler/internal/types"

type Pattern interface {
	Name() string
	Description() string
	Impact() string
	Detector() func(source string) []types.Issue
}

var registeredPatterns []Pattern

func RegisterPattern(p Pattern) {
	registeredPatterns = append(registeredPatterns, p)
}

func GetPatterns() []Pattern {
	return registeredPatterns
}
