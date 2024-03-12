package v1

import (
	"net/http"
	"testing"
)

func Test_getMetricJSONHandler_ServeHTTP(t *testing.T) {

	type args struct {
		rw http.ResponseWriter
		r  *http.Request
	}
	tests := []struct {
		name string
		h    *getMetricJSONHandler
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.h.ServeHTTP(tt.args.rw, tt.args.r)
		})
	}
}
