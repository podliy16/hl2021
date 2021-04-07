package main

import (
	"bytes"
	"compress/zlib"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/uber/jaeger-client-go"
	jaeger2 "github.com/uber/jaeger-client-go/thrift-gen/jaeger"
	"goldclient/models"
	"os"
	"time"
)

type Span struct {
	*jaeger2.Span
	TraceId   int64  `json:"traceID"`
	ProcessId string `json:"processID"`
}

type Process struct {
	ServiceName string     `json:"serviceName"`
	Tags        []struct{} `json:"tags"`
}

type Trace struct {
	TraceId   string             `json:"traceID"`
	Spans     []*Span            `json:"spans"`
	Processes map[string]Process `json:"processes"`
}

type FullTrace struct {
	Data   []Trace    `json:"data"`
	Total  int        `json:"total"`
	Limit  int        `json:"limit"`
	Offset int        `json:"offset"`
	Errors []struct{} `json:"errors"`
}

type JaegerLogger struct {
	world *models.World
}

type MySpan struct {
	StartTime int64
	Duration  int
}

func (JaegerLogger) Error(msg string) {
	fmt.Fprintf(os.Stderr, "JAGGER ERROR %s \n", msg)
}

func (j JaegerLogger) Infof(msg string, args ...interface{}) {

	//goland:noinspection GoVetCopyLock
	span := args[0].(*jaeger.Span)
	//if len(j.world.Spans) < 500 {
	j.world.Spans = append(j.world.Spans, span)
	//}
	//fmt.Fprintf(os.Stderr, "len j spans: %d \n", len(j.world.Spans))
	//
	//res, err := json.Marshal(jaeger.BuildJaegerThrift(span))
	//if err != nil {
	//	panic(err)
	//}
	//stringjson := string(res)

	//fmt.Fprintf(os.Stderr, "JAGGER NEW SPAN %s %s %s \n", msg, stringjson, j.ToJson())
}

func (j JaegerLogger) ToCustomBytes() string {
	takeSpans := 600

	b := make([]byte, takeSpans*12)
	//startTimeBuf := make([]byte, 8)
	//durationTimeBuf := make([]byte, 4)

	inc := len(j.world.Spans) / takeSpans
	indexInBuffer := 0
	//fmt.Fprintf(os.Stderr, "len spans, inc: %d %d \n",len(j.world.Spans), inc)
	for i := 0; i < len(j.world.Spans); i += inc + 1 {
		startIndex := indexInBuffer * 12
		binary.LittleEndian.PutUint64(b[startIndex:startIndex+8], uint64(j.world.Spans[i].StartTime().UnixNano()/1000))
		binary.LittleEndian.PutUint32(b[startIndex+8:startIndex+12], uint32(j.world.Spans[i].Duration().Nanoseconds()/int64(time.Microsecond)))

		indexInBuffer++
		//copy(b[startIndex:startIndex+8], startTimeBuf)
		//copy(b[startIndex+8:startIndex+12], durationTimeBuf)
	}

	return base32.StdEncoding.EncodeToString(b)
}

func (j JaegerLogger) ToJson() string {
	jaegerSpans := make([]*Span, len(j.world.Spans))
	for i := range j.world.Spans {
		tmp := jaeger.BuildJaegerThrift(j.world.Spans[i])
		jaegerSpans[i] = &Span{
			Span:      tmp,
			TraceId:   tmp.SpanId,
			ProcessId: "p1",
		}
	}

	trace := Trace{
		TraceId: "12345",
		Spans:   jaegerSpans,
		Processes: map[string]Process{
			"p1": {
				ServiceName: "goldrush",
				Tags:        []struct{}{},
			},
		},
	}

	f := FullTrace{
		Data: []Trace{
			trace,
		},
		Total:  0,
		Limit:  0,
		Offset: 0,
		Errors: nil,
	}

	res, err := json.Marshal(f)
	if err != nil {
		panic(err)
	}
	var compressed bytes.Buffer
	w, err := zlib.NewWriterLevel(&compressed, zlib.BestCompression)
	if err != nil {
		panic(err)
	}
	_, err = w.Write(res)
	if err != nil {
		panic(err)
	}
	w.Flush()
	w.Close()

	return base64.StdEncoding.EncodeToString(compressed.Bytes())
}
