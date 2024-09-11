package schedule

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/samber/lo"
)

type DataRow map[string]any

type Table struct {
	ID  string `json:"id"`
	Def struct {
		Name string `json:"name"`
	} `json:"def"`
	DataRows []DataRow `json:"data_rows"`
}

type Schedule struct {
	R struct {
		DbiAccessorRes struct {
			Tables []Table `json:"tables"`
		} `json:"DbiAccessorRes"`
	} `json:"r"`
}

type ClassDate struct {
	Name  string
	Dates []time.Time
}

// schedule is public, no authentication needed
const scheduleLocation = "https://vjg.edupage.org/timetable/server/regulartt.js?__func=regularttGetData"

type Downloader struct {
	mu sync.Mutex
	S  *Schedule
}

func (d *Downloader) GetSchedule() (*Schedule, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.S == nil {
		println("downloading schedule")
		s, err := DownloadSchedule()
		if err != nil {
			fmt.Printf("failed to download: %v\n", err)
			return nil, err
		}
		println("download complete")
		d.S = s
	}

	return d.S, nil
}

var DefaultDownloader = &Downloader{}

func DownloadSchedule() (*Schedule, error) {
	c := http.Client{Timeout: time.Second * 60}

	resp, err := c.Post(scheduleLocation, "application/json", bytes.NewBufferString(`{"__args":[null,"48"],"__gsh":"00000000"}`))
	if err != nil {
		return nil, fmt.Errorf("downloading schedule: %w", err)
	}
	s := Schedule{}

	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return nil, fmt.Errorf("reading json: %w", err)
	}
	return &s, nil
}

func GetNextClassDates(classID string, s *Schedule) ([]ClassDate, error) {
	tableByID := lo.KeyBy(s.R.DbiAccessorRes.Tables, func(item Table) string {
		return item.ID
	})

	vilniusLocation, err := time.LoadLocation("Europe/Vilnius")
	if err != nil {
		return nil, fmt.Errorf("loading time zone: %w", err)
	}

	targetClassData, found := lo.Find(tableByID["classes"].DataRows, func(item DataRow) bool {
		return item["short"] == classID
	})
	if !found {
		return nil, fmt.Errorf("class %s not found", classID)
	}

	externalClassID := targetClassData["id"].(string)

	lessons := lo.Filter(tableByID["lessons"].DataRows, func(item DataRow, index int) bool {
		for _, id := range item["classids"].([]any) {
			if id == externalClassID {
				return true
			}
		}
		return false
	})
	subjects := rowsByID(tableByID["subjects"])
	cards := rowsByID(tableByID["cards"])
	periods := rowsByID(tableByID["periods"])
	var result []ClassDate
	for _, row := range lessons {
		subj := subjects[row["subjectid"].(string)]
		dates := ClassDate{
			Name: subj["name"].(string),
		}
		for _, card := range cards {
			if card["lessonid"].(string) != row["id"].(string) {
				continue
			}
			period := periods[card["period"].(string)]

			t := time.Now().Truncate(time.Minute)
			t, err := time.ParseInLocation("2006-01-02 15:04", t.Format("2006-01-02")+" "+period["starttime"].(string), vilniusLocation)
			if err != nil {
				return nil, fmt.Errorf("parsing time: %w", err)
			}

			dates.Dates = append(dates.Dates, nextDate(t, card["days"].(string)))
		}
		slices.SortFunc(dates.Dates, func(a, b time.Time) int {
			return a.Compare(b)
		})
		result = append(result, dates)
	}
	return result, nil
}

func rowsByID(t Table) map[string]DataRow {
	return lo.KeyBy(t.DataRows, func(item DataRow) string {
		return item["id"].(string)
	})
}

func maskToWeekday(mask string) time.Weekday {
	switch mask {
	case "10000":
		return time.Monday
	case "01000":
		return time.Tuesday
	case "00100":
		return time.Wednesday
	case "00010":
		return time.Thursday
	case "00001":
		return time.Friday
	}
	panic("unknown weekday " + mask)
}

func nextDate(t time.Time, mask string) time.Time {
	result := t.AddDate(0, 0, int((maskToWeekday(mask)-t.Weekday()+7)%7))

	if result.Before(time.Now()) {
		result = result.AddDate(0, 0, 7)
	}
	return result
}

var internalNames = map[string]string{
	"Tikyba":      "Dorinis ugdymas (tikyba)",
	"1UK(An)":     "Užsienio kalba (pirmoji, anglų)",
	"Klasės val.": "Vadovavimas klasei",
	"Lietuvių k.": "Lietuvių kalba ir literatūra",
}

// ToInternalName returns corresponing discipline name in internal system.
// Some class names are mismatched between public schedule and internal class system
func ToInternalName(name string) string {
	if overide, ok := internalNames[name]; ok {
		return overide
	}
	return name
}
