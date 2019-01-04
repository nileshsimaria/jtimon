package lpserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/influxdata/influxdb/models"
)

// LPServer is Line Protocol Server
type LPServer struct {
	host string
	port int
}

// NewLPServer to create new line protocol server
func NewLPServer(host string, port int) *LPServer {
	return &LPServer{
		host: host,
		port: port,
	}
}

type lpData struct {
	sync.Mutex
	name string
	file *os.File
}

var state *lpData

func processData(db string, p string, rp string, consistency string, data []byte) error {
	state.Lock()
	defer state.Unlock()
	if state.name == "" {
		return fmt.Errorf("store is not ready")
	}

	if points, err := models.ParsePointsWithPrecision(data, time.Now().UTC(), p); err == nil {
		for _, point := range points {
			state.file.WriteString("TAGS: [\n")
			tags := point.Tags()
			for i, tag := range tags {
				state.file.WriteString(fmt.Sprintf("%d [%s = %s]\n", i, tag.Key, tag.Value))
			}
			state.file.WriteString("]\n")
			state.file.WriteString("FIELDS: [\n")
			if fields, err := point.Fields(); err == nil {
				var keys []string
				for k := range fields {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					state.file.WriteString(fmt.Sprintf("[%s=%s]\n", k, fields[k]))
				}
				state.file.WriteString("]\n")
			} else {
				return err
			}
		}
	} else {
		return err
	}

	return nil
}

func writeHandler(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("%v", err)
		return
	}
	defer r.Body.Close()

	db := r.FormValue("db")
	p := r.FormValue("precision")
	rp := r.FormValue("rp")
	consistency := r.FormValue("consistency")

	if err := processData(db, p, rp, consistency, data); err != nil {
		log.Printf("%v", err)
	}
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
	_, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}
	defer r.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, `{"status":"ok"}`)
}

type body struct {
	Name string `json:"name"`
}

func getStoreName(r *http.Request) (string, error) {
	decoder := json.NewDecoder(r.Body)
	body := &body{}
	if err := decoder.Decode(body); err != nil {
		return "", err

	}
	return body.Name, nil
}

func storeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	action := vars["op"]
	switch action {
	case "open":
		if name, err := getStoreName(r); err == nil {
			state.Lock()
			if state.name == "" {
				if state.file, err = os.Create(name); err == nil {
					state.name = name
					log.Printf("open: %s\n", name)
				} else {
					log.Printf("%v", err)
				}
			}
			state.Unlock()
		}
	case "close":
		if name, err := getStoreName(r); err == nil {
			state.Lock()
			state.Unlock()
			if state.name != "" {
				if err = state.file.Close(); err == nil {
					state.file = nil
					state.name = ""
					log.Printf("close: %s\n", name)
				} else {
					log.Printf("%v", err)
				}
			}

		}
	}
	fmt.Fprintln(w, `{"status":"ok"}`)
}

func lpInit() {
	state = &lpData{}
}

// StartServer to start the line protocol server
func (lp *LPServer) StartServer() {
	lpInit()
	r := mux.NewRouter()
	r.HandleFunc("/query", queryHandler)
	r.HandleFunc("/write", writeHandler)
	r.HandleFunc("/store/{op}", storeHandler)

	srv := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf("%s:%d", lp.host, lp.port),
	}
	log.Fatal(srv.ListenAndServe())
}
