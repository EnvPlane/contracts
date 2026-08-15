package domain

import (
	"reflect"
	"strings"
	"testing"
)

func TestJSONTagsDoNotUseOmitemptyOnStructFields(t *testing.T) {
	types := []reflect.Type{
		reflect.TypeOf(Project{}), reflect.TypeOf(Environment{}),
		reflect.TypeOf(RemoteCluster{}), reflect.TypeOf(ControlPlaneSettings{}),
		reflect.TypeOf(AuthenticationSettings{}), reflect.TypeOf(AgentResourceScanRequest{}),
	}
	for _, typ := range types {
		for index := 0; index < typ.NumField(); index++ {
			field := typ.Field(index)
			if field.Type.Kind() == reflect.Struct && strings.Contains(field.Tag.Get("json"), ",omitempty") {
				t.Fatalf("%s.%s uses omitempty on a struct field", typ.Name(), field.Name)
			}
		}
	}
}
