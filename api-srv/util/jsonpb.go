package util

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/micro/protobuf/proto"
)

var marshaler = &jsonpb.Marshaler{
	EmitDefaults: true,
}
var unmarshaler = &jsonpb.Unmarshaler{}

// Marshal marshals a protobuf message from a service call to JSON
func Marshal(pb proto.Message) json.RawMessage {
	var buf bytes.Buffer
	if err := marshaler.Marshal(&buf, pb); err != nil {
		panic(err)
	}
	return json.RawMessage(buf.Bytes())
}

// Unmarshal unmarshals a JSON-encoded protobuf message into a protobuf object
func Unmarshal(msg string, pb proto.Message) error {
	return unmarshaler.Unmarshal(strings.NewReader(msg), pb)
}
