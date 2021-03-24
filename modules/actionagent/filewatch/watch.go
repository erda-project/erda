package filewatch

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/hpcloud/tail"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/pkg/filehelper"
)

// FullHandler handle full text as io.ReadCloser
type FullHandler func(io.ReadCloser) error

// TailHandler let you can do what you want depends on currentLine and allLines
// allLines include currentLine
type TailHandler func(line string, allLines []string) error

type TailHandlerWithIO struct {
	Handler TailHandler
	TailIO  *tail.Tail
}

type Watcher struct {
	fullFileHandlerMap map[string]FullHandler
	fullFsWatcher      *fsnotify.Watcher
	errs               []error
	GracefulDoneC      chan struct{} // send when graceful stop done
}

const logPrefix = "[Platform Log] [file watcher]"

var logger = log.New(os.Stderr, logPrefix+" ", 0)

func New() (*Watcher, error) {
	w := Watcher{fullFileHandlerMap: make(map[string]FullHandler), GracefulDoneC: make(chan struct{})}

	watcher, err := fsnotify.NewWatcher()
	w.fullFsWatcher = watcher
	if err != nil {
		return nil, err
	}

	// 全量处理
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// 监听 remove 和 rename 操作，重新创建文件并监听
				if event.Op == fsnotify.Remove || event.Op == fsnotify.Rename {
					_ = filehelper.CreateFile(event.Name, "", 0644)
					_ = watcher.Add(event.Name)
					continue
				}
				// 只监听 写入 事件
				if event.Op != fsnotify.Write {
					continue
				}
				// 全量处理
				fullHandler, ok := w.fullFileHandlerMap[event.Name]
				if ok {
					content, err := ioutil.ReadFile(event.Name)
					if err != nil {
						continue
					}
					if err := fullHandler(ioutil.NopCloser(bytes.NewBuffer(content))); err != nil {
						logger.Printf("failed to handle file: %s, err: %v\n", event.Name, err)
						continue
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.Printf("ignore, watched an error: %v\n", err)
				return
			}
		}
	}()

	return &w, nil
}

func (w *Watcher) Close() {
	// max wait 10 sec
	maxWaitTime := time.Second * 10
	timer := time.NewTimer(maxWaitTime)
	select {
	case <-timer.C:
		logrus.Debugf("%s exceed max wait time (%v s) when close file watcher, force close", logPrefix, maxWaitTime.Seconds())
	case <-w.GracefulDoneC:
		logrus.Debugf("%s close file watcher success", logPrefix)
		close(w.GracefulDoneC)
	}

	if w.fullFsWatcher != nil {
		_ = w.fullFsWatcher.Close()
	}
}

func (w *Watcher) RegisterFullHandler(fullpath string, handler FullHandler) {
	w.fullFileHandlerMap[fullpath] = handler

	if _, exist := w.fullFileHandlerMap[fullpath]; exist {
		_ = w.fullFsWatcher.Remove(fullpath)
	}

	if err := w.fullFsWatcher.Add(fullpath); err != nil {
		if os.IsNotExist(err) {
			_ = filehelper.CreateFile(fullpath, "", 0644)
			_ = w.fullFsWatcher.Add(fullpath)
		} else {
			logger.Printf("ignore, failed to register full handler, fullpath: %s", fullpath)
		}
	}
}

func (w *Watcher) RegisterTailHandler(fullpath string, handler TailHandler) {
	tailIO, err := tail.TailFile(fullpath, tail.Config{ReOpen: true, MustExist: false, Follow: true, Poll: true})
	if err != nil {
		logger.Printf("ignore, failed to register tail handler, fullpath: %s, err: %v", fullpath, err)
		return
	}

	var allLines []string
	go func() {
		for line := range tailIO.Lines {
			allLines = append(allLines, line.Text)
			if err := handler(line.Text, allLines); err != nil {
				logger.Printf("failed to handle a tailed line of %s, err: %v\n", fullpath, err)
			}
		}
	}()
}
