package collector

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseLessonInfo(t *testing.T) {
	r := require.New(t)
	sampleInput := "tomval_AjaxCmd('getLessonInfo', '823bd4291bf8da3573940b3353073f41', '643344', this); return false;"
	got := parseLessonInfoCommand(sampleInput)
	r.Equal("643344", got)
}

func TestParseLessonInfoResponse(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected LessonInfo
	}{
		"basic": {
			input: `<p style="cursor:pointer;" id="closeLessonInfo" align="right" onclick="closeLessonInfo()" class='hRED'><strong>X</strong></p><b>Mokytoja(s): </b>Jona&scaron; Jonaiti&scaron;<br /><br /><b>Tema: </b>Įvadinė pamoka. ABCD.<br /><br /><b>Užduotys: </b>Perskaityti p.:6-7; užsira&scaron;yti ir pasiruo&scaron;ti sąsiuvinį darbui. Pra&scaron;au aplenkti vadovėlius. <br />Iki rugsėjo 12 d. labai pra&scaron;au nusipirkti pratybas Kelias. Užduočių sąsiuvinis 5 kl., I d.: https://www.briedis.lt/Mokyklai/5-12-Klases/Istorija/Kelias-Uzduociu-sasiuvinis-5-kl-I-d.html<br />`,
			expected: LessonInfo{
				Teacher: "Jonaš Jonaitiš",
				Topic:   "Įvadinė pamoka. ABCD.",
				Assignments: []string{
					"Perskaityti p.:6-7; užsirašyti ir pasiruošti sąsiuvinį darbui. Prašau aplenkti vadovėlius. ",
					"Iki rugsėjo 12 d. labai prašau nusipirkti pratybas Kelias. Užduočių sąsiuvinis 5 kl., I d.: https://www.briedis.lt/Mokyklai/5-12-Klases/Istorija/Kelias-Uzduociu-sasiuvinis-5-kl-I-d.html",
				},
			},
		},
		"no assignments": {
			input: `<p style="cursor:pointer;" id="closeLessonInfo" align="right" onclick="closeLessonInfo()" class='hRED'><strong>X</strong></p><b>Mokytoja(s): </b>Jonas Jonaitis<br /><br /><b>Tema: </b>Susipažinimas su taisyklėmis, vertinimu ir programa<br /><br /><b>Užduotys: </b><br />`,
			expected: LessonInfo{
				Teacher:     "Jonas Jonaitis",
				Topic:       "Susipažinimas su taisyklėmis, vertinimu ir programa",
				Assignments: nil,
			},
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			got, err := parseLessonInfoResponse(tt.input)
			r := require.New(t)
			r.NoError(err)
			r.NotNil(got)
			r.Equal(tt.expected, *got)
		})
	}
}
