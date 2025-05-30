package handlers

import (
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"go-sample/internal/models"
	"go-sample/internal/repository"
)

type ImportHandler struct {
	userRepo repository.UserRepository
	teamRepo repository.TeamRepository
	// Maximum number of concurrent file processing goroutines
	maxFileWorkers int
	// Maximum number of concurrent line processing goroutines per file
	maxLineWorkers int
}

type ImportFileRequest struct {
	EntityType string `json:"entity_type"` // "users" or "teams"
	Data       string `json:"data"`        // base64 encoded CSV
}

type ImportRequest struct {
	Files []ImportFileRequest `json:"files"`
}

type FileImportResult struct {
	EntityType    string   `json:"entity_type"`
	TotalLines    int      `json:"total_lines"`
	SuccessCount  int      `json:"success_count"`
	FailureCount  int      `json:"failure_count"`
	FailedRecords []string `json:"failed_records,omitempty"`
}

type ImportResponse struct {
	TotalFiles   int               `json:"total_files"`
	Results      []FileImportResult `json:"results"`
	ProcessingTime string           `json:"processing_time"`
}

func NewImportHandler(userRepo repository.UserRepository, teamRepo repository.TeamRepository) *ImportHandler {
	return &ImportHandler{
		userRepo:       userRepo,
		teamRepo:       teamRepo,
		maxFileWorkers: 5,                // Process up to 5 files concurrently
		maxLineWorkers: 20,               // Process up to 20 lines concurrently per file
	}
}

func (h *ImportHandler) ImportCSV(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	
	var req ImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if len(req.Files) == 0 {
		ErrorResponse(w, http.StatusBadRequest, "No files provided")
		return
	}

	// Create a channel to limit concurrent file processing
	fileWorkerCh := make(chan struct{}, h.maxFileWorkers)
	var wg sync.WaitGroup
	
	// Create a slice to store results
	var mu sync.Mutex
	results := make([]FileImportResult, 0, len(req.Files))

	// Process each file
	for i, fileReq := range req.Files {
		// Validate entity type
		if fileReq.EntityType != "users" && fileReq.EntityType != "teams" {
			result := FileImportResult{
				EntityType:    fileReq.EntityType,
				FailedRecords: []string{fmt.Sprintf("Invalid entity type: %s", fileReq.EntityType)},
				FailureCount:  1,
			}
			mu.Lock()
			results = append(results, result)
			mu.Unlock()
			continue
		}

		wg.Add(1)
		
		// Acquire a worker slot
		fileWorkerCh <- struct{}{}
		
		// Process file in a goroutine
		go func(idx int, fileRequest ImportFileRequest) {
			defer wg.Done()
			defer func() { <-fileWorkerCh }() // Release worker slot when done
			
			// Process the file and get result
			result := h.processFile(fileRequest.EntityType, fileRequest.Data)
			result.EntityType = fileRequest.EntityType
			
			// Add result to results slice
			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}(i, fileReq)
	}

	// Wait for all file processing to complete
	wg.Wait()
	close(fileWorkerCh)

	// Create response
	response := ImportResponse{
		TotalFiles:     len(req.Files),
		Results:        results,
		ProcessingTime: time.Since(startTime).String(),
	}

	SuccessResponse(w, http.StatusOK, response)
}

func (h *ImportHandler) processFile(entityType, data string) FileImportResult {
	result := FileImportResult{
		FailedRecords: []string{},
	}

	// Decode base64 data
	decodedData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		result.FailedRecords = append(result.FailedRecords, "Invalid base64 data")
		result.FailureCount++
		return result
	}

	// Parse CSV
	reader := csv.NewReader(strings.NewReader(string(decodedData)))
	
	// Read header
	header, err := reader.Read()
	if err != nil {
		result.FailedRecords = append(result.FailedRecords, "Invalid CSV format")
		result.FailureCount++
		return result
	}

	// Process CSV based on entity type
	if entityType == "users" {
		return h.processUserCSV(reader, header)
	} else {
		return h.processTeamCSV(reader, header)
	}
}

func (h *ImportHandler) processUserCSV(reader *csv.Reader, header []string) FileImportResult {
	result := FileImportResult{
		FailedRecords: []string{},
	}

	// Validate header
	requiredFields := []string{"email", "name"}
	if !validateHeader(header, requiredFields) {
		result.FailedRecords = append(result.FailedRecords, "Invalid header. Required fields: email, name")
		result.FailureCount++
		return result
	}

	// Find column indexes
	emailIdx := findColumnIndex(header, "email")
	nameIdx := findColumnIndex(header, "name")

	// Create a channel to limit concurrent line processing
	lineWorkerCh := make(chan struct{}, h.maxLineWorkers)
	var wg sync.WaitGroup
	var mu sync.Mutex // For thread-safe updates to result

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		result.FailedRecords = append(result.FailedRecords, fmt.Sprintf("Error reading CSV: %v", err))
		result.FailureCount++
		return result
	}

	result.TotalLines = len(records)

	// Process each line
	for lineNum, record := range records {
		wg.Add(1)
		
		// Acquire a worker slot
		lineWorkerCh <- struct{}{}
		
		// Process each line in a separate goroutine
		go func(record []string, lineNum int) {
			defer wg.Done()
			defer func() { <-lineWorkerCh }() // Release worker slot when done

			// Validate record
			if len(record) < len(header) {
				mu.Lock()
				result.FailedRecords = append(result.FailedRecords, 
					fmt.Sprintf("Line %d: Invalid number of fields", lineNum+1))
				result.FailureCount++
				mu.Unlock()
				return
			}

			// Create user
			user := &models.User{
				Email:     record[emailIdx],
				Name:      record[nameIdx],
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Save user
			if err := h.userRepo.Create(user); err != nil {
				mu.Lock()
				result.FailedRecords = append(result.FailedRecords, 
					fmt.Sprintf("Line %d: Failed to create user: %v", lineNum+1, err))
				result.FailureCount++
				mu.Unlock()
				return
			}

			mu.Lock()
			result.SuccessCount++
			mu.Unlock()
		}(record, lineNum)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(lineWorkerCh)

	return result
}

func (h *ImportHandler) processTeamCSV(reader *csv.Reader, header []string) FileImportResult {
	result := FileImportResult{
		FailedRecords: []string{},
	}

	// Validate header
	requiredFields := []string{"title", "description"}
	if !validateHeader(header, requiredFields) {
		result.FailedRecords = append(result.FailedRecords, "Invalid header. Required fields: title, description")
		result.FailureCount++
		return result
	}

	// Find column indexes
	titleIdx := findColumnIndex(header, "title")
	descIdx := findColumnIndex(header, "description")

	// Create a channel to limit concurrent line processing
	lineWorkerCh := make(chan struct{}, h.maxLineWorkers)
	var wg sync.WaitGroup
	var mu sync.Mutex // For thread-safe updates to result

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		result.FailedRecords = append(result.FailedRecords, fmt.Sprintf("Error reading CSV: %v", err))
		result.FailureCount++
		return result
	}

	result.TotalLines = len(records)

	// Process each line
	for lineNum, record := range records {
		wg.Add(1)
		
		// Acquire a worker slot
		lineWorkerCh <- struct{}{}
		
		// Process each line in a separate goroutine
		go func(record []string, lineNum int) {
			defer wg.Done()
			defer func() { <-lineWorkerCh }() // Release worker slot when done

			// Validate record
			if len(record) < len(header) {
				mu.Lock()
				result.FailedRecords = append(result.FailedRecords, 
					fmt.Sprintf("Line %d: Invalid number of fields", lineNum+1))
				result.FailureCount++
				mu.Unlock()
				return
			}

			// Create team
			team := &models.Team{
				Title:       record[titleIdx],
				Description: record[descIdx],
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			// Save team
			if err := h.teamRepo.Create(team); err != nil {
				mu.Lock()
				result.FailedRecords = append(result.FailedRecords, 
					fmt.Sprintf("Line %d: Failed to create team: %v", lineNum+1, err))
				result.FailureCount++
				mu.Unlock()
				return
			}

			mu.Lock()
			result.SuccessCount++
			mu.Unlock()
		}(record, lineNum)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(lineWorkerCh)

	return result
}

// Helper functions
func validateHeader(header []string, requiredFields []string) bool {
	headerMap := make(map[string]bool)
	for _, h := range header {
		headerMap[strings.ToLower(h)] = true
	}

	for _, field := range requiredFields {
		if !headerMap[field] {
			return false
		}
	}
	return true
}

func findColumnIndex(header []string, columnName string) int {
	for i, h := range header {
		if strings.ToLower(h) == columnName {
			return i
		}
	}
	return -1
} 