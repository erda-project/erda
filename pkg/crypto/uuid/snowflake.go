// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package uuid

import (
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/sony/sonyflake"
)

var sf = sonyflake.NewSonyflake(sonyflake.Settings{
	MachineID: func() (uint16, error) { return podIP(), nil },
})

// SnowFlakeIDUint64 return sequence uuid
// 39 bits for time in units of 10 msec
// 8 bits for a sequence number
// 16 bits for a machine id
func SnowFlakeIDUint64() uint64 {
	id, _ := sf.NextID()
	return id
}

// SnowFlakeID is string format SnowFlakeIDUint64
func SnowFlakeID() string {
	return strconv.FormatUint(SnowFlakeIDUint64(), 10)
}

func podIP() uint16 {
	podIP := os.Getenv("POD_IP")
	if podIP == "" {
		podIP = RandomIpV4Address()
	}
	ip := net.ParseIP(podIP)
	return uint16(ip[8])<<7 + uint16(ip[9])<<6 +
		uint16(ip[10])<<5 + uint16(ip[11])<<4 +
		uint16(ip[12])<<3 + uint16(ip[13])<<2 +
		uint16(ip[14])<<1 + uint16(ip[15])
}

// RandomIpV4Address returns a valid IPv4 address as string
func RandomIpV4Address() string {
	var blocks []string
	for i := 0; i < 4; i++ {
		number := rand.Intn(255)
		blocks = append(blocks, strconv.Itoa(number))
	}
	return strings.Join(blocks, ".")
}
