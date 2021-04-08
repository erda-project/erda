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

package writer

import (
	"fmt"
)

type testWrite struct {
	capacity int
}

func (w *testWrite) Write(data interface{}) error {
	fmt.Printf("Write data ok")
	return nil
}

func (w *testWrite) WriteN(data ...interface{}) (int, error) {
	if w.capacity <= 0 {
		err := fmt.Errorf("buffer max capacity")
		return 0, err
	}
	w.capacity -= len(data)
	fmt.Println("WriteN ", len(data), data)
	return len(data), nil
}

func (w *testWrite) Close() error {
	fmt.Println("Close")
	return nil
}

func ExampleBuffer() {
	buf := NewBuffer(&testWrite{2}, 4)
	n, err := buf.WriteN(1, 2, 3, 4, 5, 6, 7, 8, 9)
	fmt.Println(n, buf.buf, err)

	err = buf.Close()
	fmt.Println(err)

	// Output:
	// WriteN  4 [1 2 3 4]
	// 8 [5 6 7 8] buffer max capacity
	// Close
	// buffer max capacity
}

func ExampleBuffer_spitWrite() {
	buf := NewBuffer(&testWrite{10}, 3)
	n, err := buf.WriteN(1, 2, 3, 4, 5)
	fmt.Println(n, buf.buf, err)

	n, err = buf.WriteN(6)
	fmt.Println(n, buf.buf, err)

	n, err = buf.WriteN(7, 8, 9)
	fmt.Println(n, buf.buf, err)

	n, err = buf.WriteN(10, 11)
	fmt.Println(n, buf.buf, err)

	// Output:
	// WriteN  3 [1 2 3]
	// 5 [4 5] <nil>
	// WriteN  3 [4 5 6]
	// 1 [] <nil>
	// 3 [7 8 9] <nil>
	// WriteN  3 [7 8 9]
	// 2 [10 11] <nil>
}

func ExampleBuffer_for() {
	buf := NewBuffer(&testWrite{100}, 3)
	data := make([]interface{}, 10, 10)
	for i := range data {
		data[i] = i
	}
	for i := 0; i < 10; i++ {
		n, err := buf.WriteN(data[0 : 1+i%10]...)
		if err != nil {
			fmt.Println(n, buf.buf, err)
			break
		}
	}
	err := buf.Close()
	fmt.Println(buf.buf, err)

	// Output:
	// WriteN  3 [0 0 1]
	// WriteN  3 [0 1 2]
	// WriteN  3 [0 1 2]
	// WriteN  3 [3 0 1]
	// WriteN  3 [2 3 4]
	// WriteN  3 [0 1 2]
	// WriteN  3 [3 4 5]
	// WriteN  3 [0 1 2]
	// WriteN  3 [3 4 5]
	// WriteN  3 [6 0 1]
	// WriteN  3 [2 3 4]
	// WriteN  3 [5 6 7]
	// WriteN  3 [0 1 2]
	// WriteN  3 [3 4 5]
	// WriteN  3 [6 7 8]
	// WriteN  3 [0 1 2]
	// WriteN  3 [3 4 5]
	// WriteN  3 [6 7 8]
	// WriteN  1 [9]
	// Close
	// [] <nil>
}
