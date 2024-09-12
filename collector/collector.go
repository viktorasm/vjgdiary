package collector

import (
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/samber/lo"
)

const remoteLocation = "https://dienynas.vjg.lt"

type Collector struct {
	c           *colly.Collector
	loginToken  string
	StudentName string
}

func NewCollector() *Collector {
	return &Collector{
		c: colly.NewCollector(
			colly.MaxDepth(1),
		),
	}
}

func (c *Collector) WithTransport(transport http.RoundTripper) {
	c.c.WithTransport(transport)
}

func (c *Collector) Login(user string, password string) error {
	c.c.OnHTML("#top_bar > div.left.studentname > ul > li > table > tbody > tr:nth-child(1) > td:nth-child(2) > span:nth-child(1)", func(element *colly.HTMLElement) {
		c.StudentName = element.Text
	})

	const tokenSelector = "a[href^='index.php?page=login&token=']"
	c.c.OnHTML(tokenSelector, func(e *colly.HTMLElement) {
		href := e.Attr("href")

		u, err := url.Parse(href)
		if err != nil {
			return
		}
		c.loginToken = u.Query().Get("token")
		c.c.OnHTMLDetach(tokenSelector)
	})

	err := c.c.Post(remoteLocation+"/index.php?page=login&lng=&token=", map[string]string{
		"login_u": user,
		"login_p": password,
	})
	if err != nil {
		return err
	}
	if c.loginToken == "" || c.StudentName == "" {
		return fmt.Errorf("could not login")
	}

	return nil
}

func (c *Collector) GetLessonInfos() ([]*LessonInfo, error) {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	lessonsByID := map[string]*LessonInfo{}

	lessonInfoCollector := c.c.Clone()
	lessonInfoCollector.Async = true
	err := lessonInfoCollector.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 10})
	if err != nil {
		return nil, fmt.Errorf("setting up collectors: %w", err)
	}

	lessonInfoCollector.OnResponse(func(response *colly.Response) {
		resp, err := parseLessonInfoResponse(string(response.Body))
		if err != nil {
			println("FAILED TO PARSE", err.Error())
			return
		}
		lessonInfo := lessonsByID[response.Request.URL.Query().Get("id")]
		if lessonInfo != nil {
			lessonInfo.Teacher = resp.Teacher
			lessonInfo.Topic = resp.Topic
			lessonInfo.Assignments = resp.Assignments
		}
	})

	// our table is organized in lots of columns, one column per day. header tells us exact day number
	// figure out what date each column in the table represents
	c.c.OnHTML(".marks_table", func(table *colly.HTMLElement) {
		tableColumnToDate := map[int]*time.Time{}
		table.ForEach(".marks_tr_daysrow", func(i int, marksRow *colly.HTMLElement) {
			marksRow.ForEach("th[id^='m_']", func(col int, th *colly.HTMLElement) {
				// sample value: m_11_1005_: first int is month, second one is day code.
				id := th.Attr("id")
				monthText := strings.Split(id, "_")[1]
				month, err := strconv.Atoi(monthText)
				if err != nil {
					fmt.Printf("FAILED TO PARSE month/day: %s\n", monthText)
					return
				}
				th.ForEach("table.marks_table_days tr:nth-child(2)", func(i int, td *colly.HTMLElement) {
					day, err := strconv.Atoi(td.Text)
					if err != nil {
						println("FAILED TO PARSE month/day: %s", td.Text)
						return
					}
					date := time.Date(2024, time.Month(month), day, 8, 0, 0, 0, time.UTC)
					tableColumnToDate[th.DOM.Index()] = lo.ToPtr(date)
				})
			})
		})

		// now go through each row/col
		table.ForEach(".marks_tr_discrow", func(discRow int, row *colly.HTMLElement) {
			discipline := ""
			row.ForEach(".marks_td_discname", func(i int, element *colly.HTMLElement) {
				discipline = element.Text
			})
			row.ForEach("td[id^='m_']", func(col int, colContent *colly.HTMLElement) {
				colContent.ForEach(".marks_tr_markrow td[onclick^='tomval_AjaxCmd'].marks_td_markL", func(_ int, element *colly.HTMLElement) {
					lessonID := parseLessonInfoCommand(element.Attr("onclick"))
					url := fmt.Sprintf(remoteLocation+"/lessoninfo.php?time=%d&token=%s&id=%s", timestamp, c.loginToken, lessonID)

					date := tableColumnToDate[colContent.DOM.Index()]
					if date == nil {
						return
					}

					lessonInfo := LessonInfo{
						Discipline:  discipline,
						Day:         date,
						LessonNotes: parseLessonNotes(element.Attr("onmouseover")),
					}
					lessonsByID[lessonID] = &lessonInfo

					element.DOM.Find("span").Remove()
					mark := strings.TrimSpace(element.DOM.Text())
					if mark != "" {
						lessonInfo.Mark = mark
						println("mark detected:", mark)
					}

					_ = lessonInfoCollector.Visit(url)
				})
			})
		})

	})

	if err := c.c.Visit(fmt.Sprintf(remoteLocation+"/marks.php?time=%d&token=%s&semester=87&alldays=0&final=0", timestamp, c.loginToken)); err != nil {
		return nil, err
	}

	lessonInfoCollector.Wait()

	result := lo.Values(lessonsByID)
	slices.SortFunc(result, func(e *LessonInfo, e2 *LessonInfo) int {
		if e.Day == nil {
			if e2.Day == nil {
				return 0
			}
			return -1
		}
		return e.Day.Compare(*e2.Day)
	})
	return result, nil
}
