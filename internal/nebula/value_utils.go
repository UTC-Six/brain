package nebula

import (
	nebula_go "github.com/vesoft-inc/nebula-go/v3"
)

// valueWrapperToInterface converts nebula ValueWrapper to a Go native type when possible.
func valueWrapperToInterface(val nebula_go.ValueWrapper) interface{} {
	if val.IsString() {
		if v, err := val.AsString(); err == nil {
			return v
		}
	}
	if val.IsInt() {
		if v, err := val.AsInt(); err == nil {
			return v
		}
	}
	if val.IsFloat() {
		if v, err := val.AsFloat(); err == nil {
			return v
		}
	}
	if val.IsBool() {
		if v, err := val.AsBool(); err == nil {
			return v
		}
	}
	if val.IsDate() || val.IsTime() || val.IsDateTime() {
		return val.String()
	}
	if val.IsMap() || val.IsList() || val.IsSet() {
		return val.String()
	}
	return val.String()
}
