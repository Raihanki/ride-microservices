package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"ride-sharing/shared/contracts"
)

func handleTripPreview(w http.ResponseWriter, r *http.Request) {
	var request previewTripRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "fail to parse JSON data", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// validation
	if request.UserID == "" {
		http.Error(w, "userID is required", http.StatusBadRequest)
		return
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := http.Post("http://trip-service:8083/preview", "application/json", bytes.NewReader(requestBody))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "failed to call trip service: "+resp.Status, http.StatusInternalServerError)
		return
	}

	var respBody any
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		http.Error(w, "failed to parse JSON data from trip service", http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	response := contracts.APIResponse{Data: respBody}
	writeJSON(w, http.StatusCreated, response)
}
