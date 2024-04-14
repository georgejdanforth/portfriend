package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

const (
	// URL of the CSV file
	url = "https://www.iana.org/assignments/service-names-port-numbers/service-names-port-numbers.csv"
	// Local path to save the CSV file
	localPath = "ports.csv"

	minUserPort = 1024
	maxUserPort = 49151
)

// Regex to match port ranges
var rangeRegex = regexp.MustCompile(`^(\d+)-(\d+)$`)

type Port struct {
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

type PortsData struct {
	RegisteredPorts []Port
	UnregisteredPorts []uint16
}

type PortsService struct {
	mu        sync.RWMutex
	PortsData *PortsData
}

func NewPortsService() *PortsService {
	return &PortsService{
		PortsData: nil,
	}
}

func (ps *PortsService) Refresh(download bool) error {
	if !portsCsvExists() || download {
		if err := downloadPortsCsv(); err != nil {
			return err
		}
	}

	if err := ps.loadPorts(); err != nil {
		return err
	}

	return nil
}

func (ps *PortsService) loadPorts() error {
	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	registeredPorts := make([]Port, 0, len(records))
	unregisteredPorts := make([]uint16, 0)
	prevPort := uint16(0)
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

		var offset uint16

		if rangeRegex.MatchString(record[1]) {
			rangeVals := strings.Split(record[1], "-")
			start, err := strconv.ParseUint(rangeVals[0], 10, 16)
			if err != nil {
				return err
			}
			end, err := strconv.ParseUint(rangeVals[1], 10, 16)
			if err != nil {
				return err
			}

			offset = uint16(start)

			for portNumber := start; portNumber <= end; portNumber++ {
				port := Port{
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
				registeredPorts = append(registeredPorts, port)
			}
		} else {
			portNumber, err := strconv.ParseUint(record[1], 10, 16)
			if err != nil {
				log.Printf("Error parsing port number: %s at index %d", record[1], i)
				return err
			}

			offset = uint16(portNumber)

			port := Port{
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

			registeredPorts = append(registeredPorts, port)
		}

		if offset > prevPort + 1 {
			for i := prevPort + 1; i < offset; i++ {
				if i < minUserPort || i > maxUserPort {
					continue
				}
				unregisteredPorts = append(unregisteredPorts, i)
			}
		}

		prevPort = offset
	}

	ps.mu.Lock()
	ps.PortsData = &PortsData{
		RegisteredPorts: registeredPorts,
		UnregisteredPorts: unregisteredPorts,
	}
	ps.mu.Unlock()

	return nil
}

func (ps *PortsService) GetRandomUnassignedPort() (uint16, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if ps.PortsData == nil || len(ps.PortsData.UnregisteredPorts) == 0 {
		return 0, fmt.Errorf("ports data is not loaded")
	}

	index := rand.Intn(len(ps.PortsData.UnregisteredPorts))
	return ps.PortsData.UnregisteredPorts[index], nil
}

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
