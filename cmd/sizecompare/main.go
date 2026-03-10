package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"

	"github.com/tj330/bookapp/gen"
	"github.com/tj330/bookapp/metadata/pkg/model"
	"google.golang.org/protobuf/proto"
)

var metadata = &model.Metadata{
	ID:          "123",
	Title:       "Nineteen Eighty-Four",
	Description: "A classic dystopian social science fiction novel and cautionary tale about totalitarianism.",
	Author:      "George Orwell",
	ISBN:        "470015866",
}

var genMetadata = &gen.Metadata{
	Id:          "123",
	Title:       "Nineteen Eighty-Four",
	Description: "A classic dystopian social science fiction novel and cautionary tale about totalitarianism.",
	Author:      "George Orwell",
	Isbn:        "470015866",
}

func main() {
	jsonBytes, err := serializeToJson(metadata)
	if err != nil {
		panic(err)
	}

	xmlBytes, err := serializeToXML(metadata)
	if err != nil {
		panic(err)
	}

	protoBytes, err := serializeToProto(genMetadata)
	if err != nil {
		panic(err)
	}

	fmt.Printf("JSON size:\t%dB\n", len(jsonBytes))
	fmt.Printf("XML size:\t%dB\n", len(xmlBytes))
	fmt.Printf("Proto size:\t%dB\n", len(protoBytes))
}

func serializeToJson(m *model.Metadata) ([]byte, error) {
	return json.Marshal(m)
}

func serializeToXML(m *model.Metadata) ([]byte, error) {
	return xml.Marshal(m)
}

func serializeToProto(m *gen.Metadata) ([]byte, error) {
	return proto.Marshal(m)
}
