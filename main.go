package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"
)

const (
	version = "0.1"
)

var colors = map[int]string{
	0: "aqua", 1: "lime", 2: "orange", 3: "yellow", 4: "olive", 5: "silver", 6: "fuchsia", 7: "gray", 8: "green", 9: "teal", 10: "blue",
}

/*
maroon
red
orange
yellow
olive
purple
fuchsia
lime
green
navy
blue
aqua
teal
silver
gray
*/

type csvRecord struct {
	Timestamp     string
	Cluster       string
	HostnameIndex int
	Color         string
	Hostname      string
	ErrorLevel    string
	Subsystem     string
	Text          string
}

var csvRecords []csvRecord

func main() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.DebugLevel)
	log.Infof("Starting version %s, listening on port 8090", version)
	//var err error

	rtr := mux.NewRouter()
	rtr.HandleFunc("/", index)
	rtr.HandleFunc("/parse/{fname}", parse)
	srv := &http.Server{
		Handler: rtr,
		Addr:    "0.0.0.0:8090",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
func parse(w http.ResponseWriter, r *http.Request) {
	var fname string
	var ok bool
	var hostnameIndex int
	vars := mux.Vars(r)
	tmpl := template.Must(template.ParseFiles("csv_table.html"))
	hostnamesColor := make(map[string]int)

	log.Debugf("Vars=%#v", vars)

	if fname, ok = vars["fname"]; !ok {
		err := fmt.Sprintf("No file name specified: %#v", vars)
		log.Debugf(err)
		fmt.Fprintf(w, err)
	}

	csvfile, err := os.Open(fname)
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}

	// Parse the file
	csvReader := csv.NewReader(csvfile)
	//r := csv.NewReader(bufio.NewReader(csvfile))

	// Iterate through the records
	for {
		// Read each record from csv
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		record[5] = strings.Replace(record[5], "\n", "", -1)

		hostnameIndex, ok = hostnamesColor[record[2]]

		//Hostname not found, add it to the map
		if !ok {
			hostnamesColor[record[2]] = len(hostnamesColor)
			hostnameIndex = len(hostnamesColor)
		}

		csvRecords = append(csvRecords, csvRecord{
			Timestamp:     strings.Replace(record[0], "\"", "", -1),
			Cluster:       strings.Replace(record[1], "\"", "", -1),
			HostnameIndex: hostnameIndex,
			Color:         colors[hostnameIndex],
			Hostname:      strings.Replace(record[2], "\"", "", -1),
			ErrorLevel:    strings.Replace(record[3], "\"", "", -1),
			Subsystem:     strings.Replace(record[4], "\"", "", -1),
			Text:          strings.Replace(record[5], "\"", "", -1),
		})
		//log.Debugf("Records=%#v\n", csvRecords)
	}
	log.Debugf("Records=%s\n", prettyPrintInt(csvRecords))
	tmp := struct {
		Filename string
		Records  []csvRecord
	}{
		Filename: fname,
		Records:  csvRecords,
	}

	tmpl.Execute(w, tmp)

}
func index(w http.ResponseWriter, r *http.Request) {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}
	var buff bytes.Buffer
	for _, v := range files {
		name := v.Name()
		log.Debugf("V=%#v", name[0:len(name)-3])
		if name[len(name)-4:] != ".csv" {
			continue
		}

		buff.WriteString(fmt.Sprintf("<a href=\"/parse/%s\">%s</a><br>\n", name, name))

	}
	fmt.Fprintf(w, buff.String())
}
func prettyPrint(body string) string {
	var prettyJSON bytes.Buffer
	error := json.Indent(&prettyJSON, []byte(body), "", "  ")
	if error != nil {
		log.Println("JSON parse error: ", error)
		return fmt.Sprintf("%s", error)
	}
	return prettyJSON.String()
}
func prettyPrintInt(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "  ")
	return string(s)
}
