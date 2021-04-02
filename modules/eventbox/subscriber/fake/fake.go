package fake

import (
	"os"
	"time"

	"github.com/erda-project/erda/modules/eventbox/subscriber"
	"github.com/erda-project/erda/modules/eventbox/types"

	"github.com/sirupsen/logrus"
)

const (
	FakeTestFilePath = "fake_subscriber.txt"
)

type FakeSubscriber struct {
	file *os.File
}

func New(filepath string) (subscriber.Subscriber, error) {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}
	return &FakeSubscriber{file}, nil
}

func (s *FakeSubscriber) Publish(dest string, content string, timestamp int64, msg *types.Message) []error {
	time.Sleep(100 * time.Millisecond)
	logrus.Infof("FAKE: publish message: %s", string(content))
	s.file.WriteString(time.Unix(0, timestamp).Format("2006-01-02 15:04:05"))
	s.file.WriteString(" | ")
	s.file.WriteString(content)
	s.file.WriteString("\n")
	return nil
}

func (s *FakeSubscriber) Status() interface{} {
	return nil
}

func (s *FakeSubscriber) Name() string {
	return "FAKE"
}
