package main

import (
	"encoding/csv"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type PortData struct {
	ServiceName      string
	PortNumber       uint16
	Transport        string
	Description      string
	Assignee         string
	Contact          string
	RegistrationDate string
	ModificationDate string
	Reference        string
	ServiceCode      string
	UnauthorizedUse  string
	Notes            string
}

const (
	// URL of the CSV file
	url = "https://www.iana.org/assignments/service-names-port-numbers/service-names-port-numbers.csv"
	// Local path to save the CSV file
	localPath = "ports.csv"

)

// Regex to match port ranges
var rangeRegex = regexp.MustCompile(`^(\d+)-(\d+)$`)

func portsCsvExists() bool {
	_, err := os.Stat(localPath)
	return !os.IsNotExist(err)
}

func downloadPortsCsv() error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Fatalf("HTTP response code: %d", response.StatusCode)
	}

	file, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	log.Printf("File saved to %s", localPath)
	return nil
}

func loadPorts() ([]PortData, error) {
	file, err := os.Open(localPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	ports := make([]PortData, 0, len(records))
	for i, record := range records {
		if i == 0 {
			// Skip the header row
			// TODO: validate the header row
			continue
		}

		if record[1] == "" {
			// Skip rows with empty port numbers
			continue
		}

		if rangeRegex.MatchString(record[1]) {
			rangeVals := strings.Split(record[1], "-")
			start, err := strconv.ParseUint(rangeVals[0], 10, 16)
			if err != nil {
				return nil, err
			}
			end, err := strconv.ParseUint(rangeVals[1], 10, 16)
			if err != nil {
				return nil, err
			}

			for portNumber := start; portNumber <= end; portNumber++ {
				port := PortData{
					ServiceName:      record[0],
					PortNumber:       uint16(portNumber),
					Transport:        record[2],
					Description:      record[3],
					Assignee:         record[4],
					Contact:          record[5],
					RegistrationDate: record[6],
					ModificationDate: record[7],
					Reference:        record[8],
					ServiceCode:      record[9],
					UnauthorizedUse:  record[10],
					Notes:            record[11],
				}
				ports = append(ports, port)
			}
		} else {
			portNumber, err := strconv.ParseUint(record[1], 10, 16)
			if err != nil {
				log.Printf("Error parsing port number: %s at index %d", record[1], i)
				return nil, err
			}

			port := PortData{
				ServiceName:      record[0],
				PortNumber:       uint16(portNumber),
				Transport:        record[2],
				Description:      record[3],
				Assignee:         record[4],
				Contact:          record[5],
				RegistrationDate: record[6],
				ModificationDate: record[7],
				Reference:        record[8],
				ServiceCode:      record[9],
				UnauthorizedUse:  record[10],
				Notes:            record[11],
			}

			ports = append(ports, port)
		}
	}

	return ports, nil
}

func main() {
	if !portsCsvExists() {
		err := downloadPortsCsv()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Printf("File already exists at %s", localPath)
	}

	log.Println("Loading ports...")

	ports, err := loadPorts()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Loaded %d ports", len(ports))

	for i, port := range ports {
		log.Printf("%+v", port)
		if i > 100 {
			break
		}
	}
}
