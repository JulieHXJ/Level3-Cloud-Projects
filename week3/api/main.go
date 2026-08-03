package main

import (
	"encoding/json"
	"log"
	"net/http"
)



type Instance struct {
	ID		string `json:"id"`
	Name	string `json:"name"`
	Type	string `json:"type"`
	Status	string `json:"status"`
}

// local memory storage (map) for testing, search with id
var instances = make(map[string]Instance)

func main()  {
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/instances", getInstancesHandler)

	log.Println("Server is listening to http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

// r = Request
// w = Response 
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)  //status code

	_, err := w.Write([]byte(`{"status":"ok"}`)) // write into response body, returns written bytes and error
	if err != nil {
		log.Println("failed to write health response:", err)
	}


}

// return a list of instances
func getInstancesHandler(w http.ResponseWriter, r *http.Request) {
	// map -> slice
	instanceList := make([]Instance, 0, len(instances))
	
	// ignore id
	for _, instance := range instances {
		instanceList = append(instanceList, instance)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)


	// slice -> JSON array
	err := json.NewEncoder(w).Encode(instanceList)
	if err != nil {
		log.Println("failed to encode instances response:", err)
	}


}