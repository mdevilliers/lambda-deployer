package main

import (
	"encoding/json"
	"log"

	"github.com/eawsy/aws-lambda-go-core/service/lambda/runtime"
)

func main() {
	// DO NOTHING
}

func Handle(evt json.RawMessage, ctx *runtime.Context) (string, error) {
	log.Println("handle event : ", string(evt))

	return "ok", nil
}
