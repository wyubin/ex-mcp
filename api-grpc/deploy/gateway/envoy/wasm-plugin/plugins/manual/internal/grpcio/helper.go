package grpcio

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"

	"google.golang.org/protobuf/proto"
)

func extractParamkeys(pattern string) []string {
	re := regexp.MustCompile(`\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(pattern, -1)

	var params []string
	for _, m := range matches {
		params = append(params, m[1])
	}
	return params
}

// --- gRPC framing helper ---
func GrpcFrame(data []byte) []byte {
	frame := make([]byte, 5+len(data))
	frame[0] = 0 // compression flag
	binary.BigEndian.PutUint32(frame[1:5], uint32(len(data)))
	copy(frame[5:], data)
	return frame
}

// helper: 建立一個新的可寫指標實例
func NewInstance(model interface{}) interface{} {
	if model == nil {
		return nil
	}
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		return reflect.New(t.Elem()).Interface()
	}
	return reflect.New(t).Interface()
}

// shared gRPC response -> JSON helper
func GrpcBodyToJSON(grpcBody []byte, resp proto.Message) ([]byte, error) {
	if len(grpcBody) < 5 {
		return nil, errors.New("grpc body not enough")
	}
	payload := grpcBody[5:]
	if err := proto.Unmarshal(payload, resp); err != nil {
		return nil, fmt.Errorf("grpcBody encode error\n->%w", err)
	}
	return json.Marshal(resp)
}
