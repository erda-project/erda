package extmarketsvc

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/erda-project/erda/bundle"
)

func TestExtMarketSvc_SearchActions(t *testing.T) {
	os.Setenv("DICEHUB_ADDR", "dicehub.default.svc.cluster.local:10000")
	bdl := bundle.New(bundle.WithDiceHub())
	s := New(bdl)
	m, n, err := s.SearchActions([]string{"java-sec2"})
	if err != nil {
		log.Fatalln(err)
	}
	for _, v := range m {
		b, _ := json.MarshalIndent(&v, "", "  ")
		fmt.Println(string(b))
	}
	for _, v := range n {
		b, _ := json.MarshalIndent(&v, "", "  ")
		fmt.Println(string(b))
	}
}
