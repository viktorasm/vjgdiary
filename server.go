package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	fs2 "io/fs"
	"net/http"
	"slices"

	"github.com/gorilla/mux"
	"github.com/samber/lo"

	"vjgdienynas/collector"
	"vjgdienynas/schedule"
	"vjgdienynas/ui"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Name string `json:"name"`
}

func BuildServer() *mux.Router {
	// Create a new ServeMux router
	mux := mux.NewRouter()

	api := mux.PathPrefix("/api").Subrouter()

	api.HandleFunc("/login", loggedInHandler).Methods("GET")
	api.HandleFunc("/login", loginHandler).Methods("POST")
	api.HandleFunc("/logout", logoutHandler).Methods("POST")
	api.HandleFunc("/lesson-info", lessonInfoHandler).Methods("GET")

	rootDir, err := fs2.Sub(ui.Build, "build")
	if err != nil {
		panic(err)
	}
	// Serve embedded static files
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		http.ServeFileFS(writer, request, rootDir, "index.html")
	})
	fs := http.FS(rootDir)
	fileServer := http.FileServer(fs)
	mux.PathPrefix("/_app").Handler(fileServer)

	mux.PathPrefix("").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Add("Location", "/")
		writer.WriteHeader(http.StatusFound)
	})

	return mux
	//err := http.ListenAndServe(fmt.Sprintf(":%d", &cfg.Port), mux)
	//if err != nil {
	//	fmt.Println("Error starting server:", err)
	//}
}

func logoutHandler(writer http.ResponseWriter, request *http.Request) {
	// Create a cookie with the same name but set to expire in the past
	cookie := &http.Cookie{
		Name:   "login_details",
		Value:  "",  // Empty value
		Path:   "/", // Must match the path of the cookie to delete it
		MaxAge: -1,  // Expires immediately
	}

	// Set the cookie in the response to delete it
	http.SetCookie(writer, cookie)
	writer.WriteHeader(http.StatusOK)
}

func loggedInHandler(writer http.ResponseWriter, request *http.Request) {
	println("checking if logged in...")
	c := loginCollector(writer, request)
	if c == nil {
		return
	}

	respondWithJson(writer, &LoginResponse{
		Name: c.StudentName,
	})
}

func loginHandler(writer http.ResponseWriter, request *http.Request) {
	loginRequest := LoginRequest{}
	err := json.NewDecoder(request.Body).Decode(&loginRequest)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	c := collector.NewCollector()

	if err := c.Login(loginRequest.Username, loginRequest.Password); err != nil {
		http.Error(writer, err.Error(), http.StatusForbidden)
		return
	}

	cookieContent, err := json.Marshal(loginRequest)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a new cookie
	cookie := &http.Cookie{
		Name:  "login_details",
		Value: base64.StdEncoding.EncodeToString(cookieContent),
		Path:  "/",
		// Optional settings
		MaxAge:   3600,  // 1 hour
		HttpOnly: true,  // Prevent JavaScript access
		Secure:   false, // Set to true if using HTTPS
	}

	// Set the cookie in the response
	http.SetCookie(writer, cookie)

	respondWithJson(writer, &LoginResponse{
		Name: c.StudentName,
	})
}

var cachedSchedule *schedule.Schedule

func lessonInfoHandler(writer http.ResponseWriter, request *http.Request) {
	c := loginCollector(writer, request)
	if c == nil {
		return
	}

	lessons, err := c.GetLessonInfos()
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	// enrich with timing data
	if cachedSchedule == nil {
		cachedSchedule, err = schedule.DownloadSchedule()
		if err != nil {
			http.Error(writer, "could not download schedule: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := enrichLessonsWithSchedule(lessons, cachedSchedule); err != nil {
		http.Error(writer, "failed to enrich lessons with schedule: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJson(writer, lessons)
}

func respondWithJson(writer http.ResponseWriter, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(writer).Encode(value); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

func loginCollector(writer http.ResponseWriter, request *http.Request) *collector.Collector {
	c := collector.NewCollector()
	loginCookie, err := request.Cookie("login_details")
	if err != nil {
		http.Error(writer, err.Error(), http.StatusUnauthorized)
		return nil
	}
	loginCookieValue, err := base64.StdEncoding.DecodeString(loginCookie.Value)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusUnauthorized)
		return nil
	}

	loginInfo := LoginRequest{}
	if err := json.Unmarshal(loginCookieValue, &loginInfo); err != nil {
		http.Error(writer, err.Error(), http.StatusUnauthorized)
		return nil
	}

	if err := c.Login(loginInfo.Username, loginInfo.Password); err != nil {
		http.Error(writer, err.Error(), http.StatusForbidden)
		return nil
	}

	return c
}

func enrichLessonsWithSchedule(lessons []*collector.LessonInfo, s *schedule.Schedule) error {
	dates, err := schedule.GetNextClassDates("5d", s)
	if err != nil {
		return fmt.Errorf("getting class dates: %w", err)
	}
	datesByDiscipline := lo.KeyBy(dates, func(item schedule.ClassDate) string {
		return schedule.ToInternalName(item.Name)
	})

	for _, l := range lessons {
		if disciplineData, ok := datesByDiscipline[l.Discipline]; ok {
			l.NextDates = disciplineData.Dates
		} else {
			println("could not find discipline dates for", l.Discipline)
		}
	}

	slices.SortFunc(lessons, func(a, b *collector.LessonInfo) int {
		if a.NextDates == nil {
			if b.NextDates != nil {
				return -1
			}
			return 0
		}

		if b.NextDates == nil {
			return 1
		}

		return a.NextDates[0].Compare(b.NextDates[0])
	})
	return nil
}
