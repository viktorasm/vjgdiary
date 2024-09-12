package schedule

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
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
	mu       sync.Mutex
	Schedule *Schedule
	client   *http.Client
	cache    *Cache
}

func NewDownloader() (*Downloader, error) {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 40 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 40 * time.Second,
	}

	c := &http.Client{
		Timeout:   60 * time.Second, // Overall request timeout
		Transport: transport,
	}

	d := &Downloader{
		client: c,
	}

	cacheBucket := os.Getenv("CACHE_BUCKET")
	if cacheBucket != "" {
		cache, err := NewCache(cacheBucket)
		if err != nil {
			return nil, fmt.Errorf("creating cache: %w", err)
		}
		d.cache = cache
	}

	return d, nil
}

func (d *Downloader) GetSchedule(ctx context.Context) (*Schedule, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.Schedule == nil {
		err := d.restoreCache(ctx)
		if err != nil {
			return nil, fmt.Errorf("restoring cache: %w", err)
		}
	}

	if d.Schedule == nil {
		println("downloading schedule")
		s, err := d.downloadSchedule()
		if err != nil {
			fmt.Printf("failed to download: %v\n", err)
			return nil, err
		}
		println("download complete")
		d.Schedule = s

		if err := d.updateCache(ctx); err != nil {
			return nil, err
		}
	}

	return d.Schedule, nil
}

func (d *Downloader) downloadSchedule() (*Schedule, error) {

	req, err := http.NewRequest("POST", scheduleLocation, bytes.NewBufferString(`{"__args":[null,"48"],"__gsh":"00000000"}`))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0"+uuid.New().String())

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("downloading schedule: %w", err)
	}
	s := Schedule{}

	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return nil, fmt.Errorf("reading json: %w", err)
	}
	return &s, nil
}

func (d *Downloader) updateCache(ctx context.Context) error {
	if d.cache == nil {
		return nil
	}

	contents, err := json.Marshal(d.Schedule)
	if err != nil {
		return err
	}

	return d.cache.Write(ctx, "schedule.json", contents)
}

func (d *Downloader) restoreCache(ctx context.Context) error {
	if d.cache == nil {
		return nil
	}
	contents, err := d.cache.Read(ctx, "schedule.json")
	if err != nil {
		return fmt.Errorf("reading contents: %w", err)
	}
	var schedule Schedule

	if err := json.Unmarshal(contents, &schedule); err != nil {
		return fmt.Errorf("unmarshalling cache: %w", err)
	}

	d.Schedule = &schedule
	return nil
}

func GetClassDates(classID string, s *Schedule, timeFrom time.Time, timeTo time.Time) ([]ClassDate, error) {
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

			classDate := getClassDateByWeekday(t, card["days"].(string))

			dates.Dates = append(dates.Dates, extrapolateClassDates(classDate, timeFrom, timeTo)...)
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

func getClassDateByWeekday(t time.Time, mask string) time.Time {
	result := t.AddDate(0, 0, int((maskToWeekday(mask)-t.Weekday()+7)%7))

	return result
}

func extrapolateClassDates(date time.Time, from time.Time, to time.Time) []time.Time {
	// find first date that is after "from". go back first to ensure that date<from, then forward until we get first one after the period
	for date.After(from) {
		date = date.AddDate(0, 0, -7)
	}
	for !date.After(from) {
		date = date.AddDate(0, 0, 7)
	}

	var result []time.Time
	for !date.After(to) {
		result = append(result, date)
		date = date.AddDate(0, 0, 7)
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
