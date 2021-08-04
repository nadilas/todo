package testutils

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"log"
	"path"
	"runtime"
	"strconv"
	"time"
)

type KeyVal struct {
	Key   string
	Value interface{}
}

func KV(key string, val interface{}) KeyVal {
	return KeyVal{Key: key, Value: val}
}

func ElementsMatching(sa []string) func(a []string) bool {
	return func(a []string) bool {
		for _, m := range sa {
			if !StrArrContains(a, m) {
				return false
			}
		}
		return true
	}
}

func StrArrContains(sa []string, s1 string) bool {
	for _, s2 := range sa {
		if s2 == s1 {
			return true
		}
	}
	return false
}

func Trace(skip int) {
	callerPtr, file, line, ok := runtime.Caller(skip)
	if !ok {
		return
	}
	caller := runtime.FuncForPC(callerPtr)
	log.Println("Trace", path.Base(caller.Name()), path.Base(file)+":"+strconv.Itoa(line))
}

func TimeUntilMidnight(from time.Time, timezone *time.Location) time.Duration {
	fromInTz := from.In(timezone)
	midnight := time.Date(fromInTz.Year(), fromInTz.Month(), fromInTz.Day()+1, 0, 0, 0, 0, timezone)
	return midnight.Sub(fromInTz)
}

func MustMarshalAny(m proto.Message) *anypb.Any {
	d, _ := anypb.New(m)
	return d
}
