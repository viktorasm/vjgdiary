package collector

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"

	"github.com/samber/lo"
)

var lessonInfoRegexp = regexp.MustCompile(`tomval_AjaxCmd\('(\w+)', '(\w+)', '(\w+)', this\); return false`)

type LessonInfo struct {
	Discipline  string      `json:"discipline,omitempty"`
	Day         *time.Time  `json:"day,omitempty"`
	Teacher     string      `json:"teacher,omitempty"`
	Topic       string      `json:"topic,omitempty"`
	Assignments []string    `json:"assignments,omitempty"`
	NextDates   []time.Time `json:"nextDates,omitempty"`
}

// parseLessonInfoCommand takes an onclick handler value and extracts lesson ID from it
func parseLessonInfoCommand(input string) string {
	m := lessonInfoRegexp.FindStringSubmatch(input)
	if len(m) != 4 {
		return ""
	}
	return m[3]
}

// parseLessonInfoResponse takes a raw HTLM that is of a crappy BR separated plain format (no headers etc)
// and attempts to parse metadata from it.
func parseLessonInfoResponse(input string) (*LessonInfo, error) {
	infoStart := "<b>Mokytoja(s)"

	input = strings.ReplaceAll(input, "&scaron;", "š")
	input = strings.ReplaceAll(input, "&Scaron;", "Š")

	i := strings.Index(input, infoStart)
	if i < 0 {
		return nil, fmt.Errorf("could not parse lesson info")
	}

	input = input[i:]

	elements := strings.Split(input, "<br />")

	const teacherPrefix = `<b>Mokytoja(s): </b>`
	const topicPrefix = `<b>Tema: </b>`
	const assignmentPrefix = `<b>Užduotys: </b>`
	result := LessonInfo{}
	for _, element := range elements {
		if strings.TrimSpace(element) == "" {
			continue
		}
		if element, ok := strings.CutPrefix(element, teacherPrefix); ok {
			result.Teacher = html.UnescapeString(element)
			continue
		}
		if element, ok := strings.CutPrefix(element, topicPrefix); ok {
			result.Topic = html.UnescapeString(element)
			continue
		}

		element := strings.TrimPrefix(element, assignmentPrefix)
		result.Assignments = append(result.Assignments, html.UnescapeString(element))
	}

	result.Assignments = lo.Filter(result.Assignments, func(item string, index int) bool {
		return strings.TrimSpace(item) != ""
	})
	if len(result.Assignments) == 0 {
		result.Assignments = nil
	}

	return &result, nil
}
