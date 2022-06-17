package main

import (
	"encoding/json"
	"log"
	"time"
)

type Req struct {
	Ts time.Time `json:"ts"`
}

func main() {
	in := `{"ts": "2022-06-04T00:00:00Z"}`
	var req Req
	if err := json.Unmarshal([]byte(in), &req); err != nil {
		log.Fatal(err)
	}
	//t, err := time.Parse("2006-01-02", in)
	//if err != nil {
	//	log.Fatal(err)
	//}
	log.Printf("time: %v", req)
}
