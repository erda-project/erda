package footnote

import (
	"container/list"
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSeqFlatten(t *testing.T) {
	l := list.New()
	seq := concat(str("啊啊啊"), indent(concat(newline(), num(111), indent(concat(newline(), str("deuuiu"))))))
	l.PushFront(seqWithColumn{seq, 0})
	r := string(flatten(l, 0))
	assert.Equal(t, `啊啊啊
   111
      deuuiu`, r)
}

func TestAlignLines(t *testing.T) {
	a := map[int]string{1: "asded", 2: "uuu", 3: "iiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiii"}
	alignLines(a)
	assert.Equal(t, len(a[1]), len(a[2]))
	assert.Equal(t, len(a[1]), len(a[3]))

}

func TestNoteRegex(t *testing.T) {
	fmt.Printf("%+v\n", New("iii\nfe\n   fe\n").NoteRegex(regexp.MustCompile("(?m)^[[:blank:]]+fe"), "nn").Dump()) // debug print

}
