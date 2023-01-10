package main

import (
	"strconv"
	"testing"
)

func Test_isBogusCoinAddress(t *testing.T) {
	tests := []struct {
		address string
		want    bool
	}{
		{address: "7F1u3wSD5RbOHQmupo9nx4TnhQ", want: true},
		{address: "7iKDZEwPZSqIvDnHvVN2r0hUWXD5rHX", want: true},
		{address: "7LOrwbDlS8NujgjddyogWgIM93MV5N2VR", want: true},
		{address: "7adNeSwJkMakpEcln9HEtthSRtxdmEHOT8T", want: true},
		{address: "7adNeSwJkMakpEcln9HEtthSRtxdmEHOT8Tasdasdasdsadasdsad", want: false},
		{address: "2F1u3wSD5RbOHQmupo9nx4TnhQ", want: false},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if got := isBogusCoinAddress(tt.address); got != tt.want {
				t.Errorf("isBogusCoinAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}
