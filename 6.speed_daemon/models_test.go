package main

import (
	"bytes"
	"reflect"
	"strconv"
	"testing"
)

func TestErrMsg_Encode(t *testing.T) {
	tests := []struct {
		m    Response
		want []byte
	}{
		{m: Response{ErrMsg: "bad", Type: Error}, want: []byte{0x10, 0x03, 0x62, 0x61, 0x64}},
		{m: Response{ErrMsg: "illegal msg", Type: Error}, want: []byte{0x10, 0x0b, 0x69, 0x6c, 0x6c, 0x65, 0x67, 0x61, 0x6c, 0x20, 0x6d, 0x73, 0x67}},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if got := tt.m.Encode(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ErrMsg.Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequest_Decode(t *testing.T) {
	tests := []struct {
		input []byte
		want  Request
	}{
		{want: Request{Plate: PlateData{name: "UN1X", timestamp: 1000}, Type: Plate}, input: []byte{0x20, 0x04, 0x55, 0x4e, 0x31, 0x58, 0x00, 0x00, 0x03, 0xe8}},
		{want: Request{Plate: PlateData{name: "UN1X", timestamp: 1000}, Type: Plate}, input: []byte{0x20, 0x04}}, // Incomplete bytes
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			buf := bytes.NewBuffer(tt.input)
			if got, _ := DecodeRequest(buf); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Request.Decode() = %v, want %v", got, tt.want)
			}
		})
	}
}
