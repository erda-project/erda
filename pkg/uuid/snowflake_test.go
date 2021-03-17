package uuid

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/sony/sonyflake"
)

func TestSnowFlakeUUID(t *testing.T) {
	fmt.Println(SnowFlakeIDUint64())
	fmt.Println(SnowFlakeIDUint64())
	fmt.Println(SnowFlakeIDUint64())
	fmt.Println(SnowFlakeIDUint64())
}

func TestSnowFlakeUUIDConcurrency(t *testing.T) {
	// different pod ip cause not duplicated
	sf1 := sonyflake.NewSonyflake(sonyflake.Settings{
		MachineID: func() (uint16, error) { return podIP(), nil },
	})
	sf2 := sonyflake.NewSonyflake(sonyflake.Settings{
		MachineID: func() (uint16, error) { return podIP(), nil },
	})
	var sf1Results []string
	var sf2Results []string
	sf1Done := make(chan struct{})
	sf2Done := make(chan struct{})
	go func() {
		for i := 0; i < 10000; i++ {
			id1, _ := sf1.NextID()
			sf1Results = append(sf1Results, strconv.FormatUint(id1, 10))
		}
		sf1Done <- struct{}{}
	}()
	go func() {
		for i := 0; i < 10000; i++ {
			id2, _ := sf2.NextID()
			sf2Results = append(sf2Results, strconv.FormatUint(id2, 10))
		}
		sf2Done <- struct{}{}
	}()
	<-sf1Done
	<-sf2Done
	sf1Map := make(map[string]struct{})
	for _, r1 := range sf1Results {
		sf1Map[r1] = struct{}{}
	}
	for _, r2 := range sf2Results {
		if _, ok := sf1Map[r2]; ok {
			panic(fmt.Errorf("id %s duplicated", r2))
		}
	}
}

func TestSnowFlakeUUIDConcurrencyNotOK(t *testing.T) {
	// no machine id cause duplicated when distributed
	sf1 := sonyflake.NewSonyflake(sonyflake.Settings{
		//MachineID: func() (uint16, error) { return podIP(), nil },
	})
	sf2 := sonyflake.NewSonyflake(sonyflake.Settings{
		//MachineID: func() (uint16, error) { return podIP(), nil },
	})
	var sf1Results []string
	var sf2Results []string
	sf1Done := make(chan struct{})
	sf2Done := make(chan struct{})
	go func() {
		for i := 0; i < 10000; i++ {
			id1, _ := sf1.NextID()
			sf1Results = append(sf1Results, strconv.FormatUint(id1, 10))
		}
		sf1Done <- struct{}{}
	}()
	go func() {
		for i := 0; i < 10000; i++ {
			id2, _ := sf2.NextID()
			sf2Results = append(sf2Results, strconv.FormatUint(id2, 10))
		}
		sf2Done <- struct{}{}
	}()
	<-sf1Done
	<-sf2Done
	sf1Map := make(map[string]struct{})
	for _, r1 := range sf1Results {
		sf1Map[r1] = struct{}{}
	}
	for _, r2 := range sf2Results {
		if _, ok := sf1Map[r2]; ok {
			panic(fmt.Errorf("id %s duplicated", r2))
		}
	}
}
