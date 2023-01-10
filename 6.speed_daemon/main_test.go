package main

import (
	"reflect"
	"strconv"
	"testing"
)

func Test_detectOverSpeeding(t *testing.T) {
	type args struct {
		platesOfRoad []PlateRecord
		road         uint16
		plate        string
		speedLimit   uint16
	}

	tests := []struct {
		args args
		want []TicketMsg
	}{
		{
			args: args{
				plate:      "HU09VPF",
				road:       14931,
				speedLimit: 60,
				platesOfRoad: []PlateRecord{
					{mile: 1120, timestamp: 56264963},
					{mile: 10, timestamp: 56326797},
				},
			}, want: []TicketMsg{
				{Plate: "HU09VPF", Mile1: 1120, Timestamp1: 56264963, Mile2: 10, Timestamp2: 56326797, Road: 14931, Speed: 6462},
			},
		},
		{
			args: args{
				plate:      "UN1X",
				road:       120,
				speedLimit: 20,
				platesOfRoad: []PlateRecord{
					{mile: 10, timestamp: 56370000},
					{mile: 100, timestamp: 56380000},
					{mile: 200, timestamp: 56390000},
				},
			}, want: []TicketMsg{
				{Plate: "UN1X", Mile1: 10, Timestamp1: 56370000, Mile2: 100, Timestamp2: 56380000, Road: 120, Speed: 3240},
				{Plate: "UN1X", Mile1: 100, Timestamp1: 56380000, Mile2: 200, Timestamp2: 56390000, Road: 120, Speed: 3600},
			},
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if got := detectOverSpeeding(tt.args.road, tt.args.speedLimit, tt.args.plate, tt.args.platesOfRoad); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("detectOverSpeeding() = %v, want %v", got, tt.want)
			}
		})
	}
}
