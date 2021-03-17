package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"strings"
)

const RefKey = "$ref"

func main() {
	var input, output string
	flag.StringVar(&input, "if", "", "input file")
	flag.StringVar(&output, "of", "", "output file")

	flag.Parse()

	if input == "" || output == "" {
		log.Fatal("no file specified, input", input, ", output", output)
	}

	b, err := ioutil.ReadFile(input)
	if err != nil {
		log.Fatal(err)
	}

	data := make(map[string]interface{})
	if err := json.Unmarshal(b, &data); err != nil {
		log.Fatal(err)
	}

	definitions, ok := data["definitions"]
	if !ok {
		log.Fatal("no definitions")
	}

	defs := definitions.(map[string]interface{})
	for k, v := range defs {
		vMap := v.(map[string]interface{})
		refKey, ok := vMap[RefKey]
		if !ok {
			continue
		}

		if strings.TrimPrefix(refKey.(string), "#/definitions/") == k {
			log.Println("removing circular ref key", refKey)
			delete(vMap, RefKey)
		}
	}

	out, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(output, out, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
