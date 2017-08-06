package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type Session struct {
	ID         int    `json:"id,omitempty"`
	Create     string `json:"Created At,omitempty"`
	Duration   int    `json:"Duration,omitempty"`
	Datapoints []dataPoint
}

type Person struct {
	ID       string `json:"id,omitempty"`
	Create   string `json:"Created At,omitempty"`
	Username string `json:"Username,omitempty"`
	Gender   string `json:"Gender,omitempty"`
	Age      int    `json:"Age,omitempty"`
	HR1L     int    `json:"HR zone1 Lower,omitempty"`
	HR1U     int    `json:"HR zone1 Upper,omitempty"`
	HR2L     int    `json:"HR zone2 Lower,omitempty"`
	HR2U     int    `json:"HR zone2 Upper,omitempty"`
	HR3L     int    `json:"HR zone3 Lower,omitempty"`
	HR3U     int    `json:"HR zone3 Upper,omitempty"`
	HR4L     int    `json:"HR zone4 Lower,omitempty"`
	HR4U     int    `json:"HR zone4 Upper,omitempty"`
	userSes  []int
}

type dataPoint struct {
	SessionID int    `json:"Session ID,omitempty"`
	BPM       int    `json:"BPM,omitempty"`
	StartTime string `json:"Start Time,omitempty"`
	EndTime   string `json:"End Time,omitempty"`
	Duration  int    `json:"Duration,omitempty"`
}

type HRdata struct {
	Min     int `json:"minimum,omitempty"`
	Max     int `json:"maximum,omitempty"`
	Average int `json:"average,omitempty"`
}

type zonePercent struct {
	Time1 int `json:"Time in Seconds Zone 1,omitempty"`
	Time2 int `json:"Time in Seconds Zone 2,omitempty"`
	Time3 int `json:"Time in Seconds Zone 3,omitempty"`
	Time4 int `json:"Time in Seconds Zone 4,omitempty"`
	Zone1 int `json:"Percent Zone 1,omitempty"`
	Zone2 int `json:"Percent Zone 2,omitempty"`
	Zone3 int `json:"Percent Zone 3,omitempty"`
	Zone4 int `json:"Percent Zone 4,omitempty"`
}

type fullWorkout struct {
	HR   HRdata
	Zone zonePercent
}

//defining globals
var people []Person
var sessions []Session
var sesmap map[int]int

//get userid from session id
// Preproccessing all csv files into memory for improved performance post startup
func preprocessUsers() {
	csv_data_points, err := os.Open("./users.csv")
	if err != nil {
		fmt.Println(err)
	}
	r := csv.NewReader(csv_data_points)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		age, _ := strconv.Atoi(record[4])
		hr1l, _ := strconv.Atoi(record[5])
		hr1u, _ := strconv.Atoi(record[6])
		hr2l, _ := strconv.Atoi(record[7])
		hr2u, _ := strconv.Atoi(record[8])
		hr3l, _ := strconv.Atoi(record[9])
		hr3u, _ := strconv.Atoi(record[10])
		hr4l, _ := strconv.Atoi(record[11])
		hr4u, _ := strconv.Atoi(record[12])
		newSes := make([]int, 0)
		newUser := Person{record[0], record[1], record[2], record[3], age, hr1l, hr1u, hr2l, hr2u, hr3l, hr3u, hr4l, hr4u, newSes}
		people = append(people, newUser)
	}
	csv_data_points.Close()
}

func preprocessSessions() {
	csvSessions, err := os.Open("./hrm_sessions.csv")
	if err != nil {
		fmt.Println(err)
	}
	r := csv.NewReader(csvSessions)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		sesid, err := strconv.Atoi(record[0])
		if err != nil {
			//fmt.Println("NaN, skipping first line in file")
		} else {
			uid, _ := strconv.Atoi(record[1])
			dur, _ := strconv.Atoi(record[3])
			sesmap[sesid] = uid
			newData := make([]dataPoint, 0)
			newSes := Session{sesid, record[2], dur, newData}
			sessions = append(sessions, newSes)
			people[uid].userSes = append(people[uid].userSes, sesid)
		}
	}
	csvSessions.Close()
}

func preprocessData() {
	csvData, err := os.Open("./hrm_data_points.csv")
	if err != nil {
		fmt.Println(err)
	}
	r := csv.NewReader(csvData)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		sesid, err := strconv.Atoi(record[0])
		if err != nil {
			//fmt.Println("NaN, skipping first line of file")
		} else {
			bpm, _ := strconv.Atoi(record[1])
			start := record[2]
			stop := record[3]
			dur, _ := strconv.Atoi(record[4])

			//			uid := sesmap[sesid]
			newData := dataPoint{sesid, bpm, start, stop, dur}
			//fmt.Println("sesid is:", sesid)
			sessions[sesid].Datapoints = append(sessions[sesid].Datapoints, newData)
		}
	}
	csvData.Close()
}

//Done Preproccessing
//
//Begining fuctions for calculating data
func HRMbySession(id string) HRdata {
	sesidnum, _ := strconv.Atoi(id)
	min, max, sum := 0, 0, 0
	for i := 0; i < len(sessions[sesidnum].Datapoints); i++ {
		if i == 0 {
			min = sessions[sesidnum].Datapoints[i].BPM
			max = sessions[sesidnum].Datapoints[i].BPM
		} else {
			if sessions[sesidnum].Datapoints[i].BPM < min {
				min = sessions[sesidnum].Datapoints[i].BPM
			}
			if sessions[sesidnum].Datapoints[i].BPM > max {
				max = sessions[sesidnum].Datapoints[i].BPM
			}
		}
		sum += sessions[sesidnum].Datapoints[i].BPM
	}
	var hrstruct HRdata
	if len(sessions[sesidnum].Datapoints) == 0 {
		hrstruct = HRdata{Min: min, Max: max, Average: sum}
	} else {
		hrstruct = HRdata{Min: min, Max: max, Average: sum / len(sessions[sesidnum].Datapoints)}
	}
	return hrstruct
}

func getHRzone(p Person, bpm int) int {
	if p.HR1L <= bpm && p.HR1U >= bpm {
		return 1
	} else if p.HR2L <= bpm && p.HR2U >= bpm {
		return 2
	} else if p.HR3L <= bpm && p.HR3U >= bpm {
		return 3
	} else if p.HR4L <= bpm && p.HR4U >= bpm {
		return 4
	} else {
		//fmt.Println("Outside of heart rate zones, something is wrong")
		//fmt.Println("bpm is", bpm, "and person id is", p.ID)
		return -1
	}
}

func getPecentageinZones(ses string) zonePercent {
	sesid, _ := strconv.Atoi(ses)
	s := sessions[sesid]
	time1, time2, time3, time4 := 0, 0, 0, 0
	curper := people[sesmap[s.ID]]
	for i := 0; i < len(s.Datapoints); i++ {
		//fmt.Println("i is", i, "and duration is", s.Datapoints[i].Duration)
		zone := getHRzone(curper, s.Datapoints[i].BPM)
		//fmt.Println("zone is", zone)
		if zone == 1 {
			time1 += s.Datapoints[i].Duration
		} else if zone == 2 {
			time2 += s.Datapoints[i].Duration
		} else if zone == 3 {
			time3 += s.Datapoints[i].Duration
		} else if zone == 4 {
			time4 += s.Datapoints[i].Duration
		} else {
			//fmt.Println("outside of hrzones, something is wrong")
		}
	}
	//fmt.Println("zone durations are", time1, time2, time3, time4)
	//fmt.Println("total durration is", s.Duration)
	if s.Duration == 0 {
		perzones := zonePercent{0, 0, 0, 0, 0, 0, 0, 0}
		return perzones
	} else {
		zone1 := ((time1 * 100) / s.Duration)
		zone2 := ((time2 * 100) / s.Duration)
		zone3 := ((time3 * 100) / s.Duration)
		zone4 := ((time4 * 100) / s.Duration)
		perzones := zonePercent{time1, time2, time3, time4, zone1, zone2, zone3, zone4}
		return perzones
	}
}

func GetSessionsNewestFirst(numResp int) []fullWorkout {
	allWorkouts := make([]fullWorkout, 0)
	for i := len(sessions) - 1; i > len(sessions)-(numResp+1); i-- {
		sesid := strconv.Itoa(i)
		hrinfo := HRMbySession(sesid)
		zones := getPecentageinZones(sesid)
		thisWorkout := fullWorkout{hrinfo, zones}
		//fmt.Println(thisWorkout.HR.Average)
		allWorkouts = append(allWorkouts, thisWorkout)
	}
	return allWorkouts
}

//Done with fuctions
//
//Begin API's

func GetPersonEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	for _, item := range people {
		if item.ID == params["id"] {
			json.NewEncoder(w).Encode(item)
			return
		}
	}
	json.NewEncoder(w).Encode(&Person{})
}

func GetPeopleEndpoint(w http.ResponseWriter, req *http.Request) {
	json.NewEncoder(w).Encode(people)
}

func GetAllHRM(w http.ResponseWriter, req *http.Request) {
	allBPMs := make([]HRdata, 0)
	for i := 1; i < len(sessions); i++ {
		curses := strconv.Itoa(i)
		curHRdata := HRMbySession(curses)
		allBPMs = append(allBPMs, curHRdata)
	}
	json.NewEncoder(w).Encode(allBPMs)
}

func GetSessionHRM(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	hrstruct := HRMbySession(params["id"])
	json.NewEncoder(w).Encode(hrstruct)
}

func GetSessionZones(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	zones := getPecentageinZones(params["id"])
	json.NewEncoder(w).Encode(zones)
}

func GetSessionsNewest(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	num, _ := strconv.Atoi(params["num"])
	allSesInfo := GetSessionsNewestFirst(num)
	json.NewEncoder(w).Encode(allSesInfo)
}

func main() {
	start := time.Now()
	sesmap = make(map[int]int)
	preprocessUsers()
	fmt.Println("User processing complete")
	sessions = append(sessions, Session{0, "Placeholder for off by 1", 0, nil}) //session ids start at 1
	preprocessSessions()
	fmt.Println("Session processing complete")
	preprocessData()
	elapsed := time.Since(start)
	fmt.Printf("Preproccessing completed in %s", elapsed)
	//	fmt.Println(len(people))
	router := mux.NewRouter()
	router.HandleFunc("/people", GetPeopleEndpoint).Methods("GET")
	router.HandleFunc("/people/{id}", GetPersonEndpoint).Methods("GET")
	//router.HandleFunc("/people/{id}/AllHRM", GetPersonHRM).Methods("GET")
	router.HandleFunc("/session/{id}/AllHRM", GetSessionHRM).Methods("GET")
	router.HandleFunc("/session/AllHRM", GetAllHRM).Methods("GET")
	router.HandleFunc("/session/{id}/zones", GetSessionZones).Methods("GET")
	router.HandleFunc("/session/HRZones/newest/{num}", GetSessionsNewest).Methods("GET")
	log.Fatal(http.ListenAndServe(":12345", router))
	return
}
