package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Weather struct {
	Temp float64 `json:"temp"`
	City string  `json:"city"`
}


func main() {
	http.HandleFunc("/weather", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Weather{Temp: 12.5, City: "Beijing"})
	})

	fmt.Println("🌤️  Server running at :8080")
	http.ListenAndServe(":8080", nil)
}
