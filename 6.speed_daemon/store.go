package main

import (
	"log"
	"sort"
)

type PlateRecord struct {
	mile      uint16
	timestamp uint32
}

// PlateRecords hold place records per plate
type PlateRecords map[string][]PlateRecord

type Store struct {
	roadPlates  map[uint16]PlateRecords // plates are tied to road
	speedLimits map[uint16]uint16
}

func newStore() *Store {
	return &Store{
		roadPlates:  make(map[uint16]PlateRecords),
		speedLimits: make(map[uint16]uint16),
	}
}

func (t *Store) SavePlate(road, mile uint16, timestamp uint32, plateName string) {
	if _, ok := t.roadPlates[road]; !ok {
		t.roadPlates[road] = make(PlateRecords)
	}

	t.roadPlates[road][plateName] = append(t.roadPlates[road][plateName], PlateRecord{mile: mile, timestamp: timestamp})
	log.Printf("Saved plate info [plateName=%s] [road=%d mile=%d timestamp=%d]\n", plateName, road, mile, timestamp)

	sort.Slice(t.roadPlates[road][plateName], func(i, j int) bool {
		return t.roadPlates[road][plateName][i].timestamp < t.roadPlates[road][plateName][j].timestamp
	})
}

func (t *Store) PlatesOfRoad(road uint16) PlateRecords {
	return t.roadPlates[road]
}

func (t *Store) SaveSpeedLimit(road, limit uint16) {
	t.speedLimits[road] = limit
}

func (t *Store) SpeedLimit(road uint16) uint16 {
	l, ok := t.speedLimits[road]
	if !ok {
		panic("Speed limit unknown for road")
	}
	return l
}
