package util

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

func Register[Input any, Output any](fn func(input *Input) Output) {
	path := FunctionPath(fn)
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		defer HandleEndpointPanic(w, r)
		in := JsonBytesToStruct[Input](r)
		out := fn(in)
		bytes := StructToJsonBytes(out)
		w.Write(bytes)
	})
}

func HandleEndpointPanic(w http.ResponseWriter, r *http.Request) {
	if r := recover(); r != nil {
		log.Println(r)
		str, isStr := r.(string)
		if isStr {
			w.WriteHeader(500)
			w.Write([]byte(str))
		}
		err, isErr := r.(error)
		if isErr {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
		}
	}
}

func ReadAll(r io.Reader) []byte {
	return Must(io.ReadAll(r))
}

func JsonBytesToStruct[T any](r *http.Request) *T {
	var t T
	bytes := ReadAll(r.Body)
	if len(bytes) == 0 {
		return nil
	}
	Check(json.Unmarshal(bytes, &t))
	defer r.Body.Close()
	return &t
}

func ParseResponse[T any](resp *http.Response) T {
	var t T
	json.Unmarshal(ReadAll(resp.Body), &t)
	defer resp.Body.Close()
	return t
}

func StructToJsonBytes(v any) []byte {
	return Must(json.Marshal(v))
}

func FunctionPath(fn any) string {
	funcName := runtime.FuncForPC((reflect.ValueOf(fn).Pointer())).Name() // "pkg/FunctionName"
	parts := strings.Split(funcName, ".")                                 // [pkg/util, FunctionName]
	if len(parts) != 2 {
		panic("issue with function name")
	}
	funcName = parts[1]
	return fmt.Sprintf("/%s", funcName)
}
