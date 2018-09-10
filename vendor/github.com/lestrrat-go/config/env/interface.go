package env

import "os"

type Source interface {
	LookupEnv(string) (string, bool)
}

type SourceFunc func(string) (string, bool)

var System Source = SourceFunc(os.LookupEnv)

type Unmarshaler interface {
	UnmarshalEnv(string) error
}

type Decoder struct {
	prefix string
	src    Source
	sep    string
}

type separatorKey struct{}
type prefixKey struct{}
