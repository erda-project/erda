// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package command

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"

	"github.com/erda-project/erda/modules/pkg/colonyutil"
	"github.com/erda-project/erda/modules/soldier/settings"
)

type GetDumpRequest struct {
	User     string `json:"user"`
	Host     string `json:"host"`
	App      string `json:"app"`
	Docker   string `json:"docker"`
	FilePath string `json:"filePath"`
	Option   string `json:"option"`
}

type Download struct {
	GetDumpRequest
	Key string
	DB  *leveldb.DB
}

type GetDumpResponse struct {
	Key         string `json:"key"`
	Description string `json:"description"`
}

type DumpResult struct {
	Dump     int `json:"dump"`
	Copy     int `json:"copy"`
	Compress int `json:"compress"`
	Upload   int `json:"upload"`
}

type DumpError struct {
	Dump     string `json:"dump"`
	Copy     string `json:"copy"`
	Compress string `json:"compress"`
	Upload   string `json:"upload"`
}

type GetDumpStatusResponse struct {
	Data  *DumpResult `json:"data"`
	Error *DumpError  `json:"error"`
}

const (
	dumpPostfix     = ".dump"
	postfix         = ".tar.gz"
	tmp             = "/tmp/"
	endpoint        = "oss-cn-hangzhou.aliyuncs.com"
	accessKeyId     = "LTAIH0Ebt2Vdi9Kf"
	accessKeySecret = "h1HpanNIR3eQjRHoRDLsK3OtVAbtt1"
	dError          = "-error"
	state           = "-status"
	success         = 1
	failed          = 2
)

func runCommand(cmd *exec.Cmd) (log string, status int) {
	cmd.Env = os.Environ()
	var stdoutBuf, stderrBuf bytes.Buffer

	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()

	var errStdout, errStderr error

	stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
	stderr := io.MultiWriter(os.Stderr, &stderrBuf)
	err := cmd.Start()
	if err != nil {
		logrus.Errorf("cmd.Start() failed with '%s'\n", err)
	}
	go func() {
		_, errStdout = io.Copy(stdout, stdoutIn)
	}()
	go func() {
		_, errStderr = io.Copy(stderr, stderrIn)
	}()

	err = cmd.Wait()

	if errStdout != nil || errStderr != nil {
		logrus.Error("failed to capture stdout or stderr\n")
	}
	errStr := string(stderrBuf.Bytes())
	stdStr := string(stdoutBuf.Bytes())

	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			logrus.Errorf("Exit Status: %d", status.ExitStatus())
			return errStr, status.ExitStatus()
		}
	}

	return stdStr, 0

}

func (d *Download) writeDB(error *DumpError, result *DumpResult) error {
	eb, err := json.Marshal(error)
	if err != nil {
		jsonError := fmt.Sprintf("Marshal %s dump error failed", d.Key)
		logrus.Errorf(jsonError)
		return errors.New(jsonError)
	}
	rb, err := json.Marshal(result)
	if err != nil {
		jsonError := fmt.Sprintf("Marshal %s dump error failed", d.Key)
		logrus.Errorf("Marshal %s dump result failed", d.Key)
		return errors.New(jsonError)
	}
	err = d.DB.Put([]byte(d.Key+dError), []byte(eb), nil)
	if err != nil {
		dbError := fmt.Sprintf("%s can not put value", d.Key+dError)
		logrus.Errorf(dbError)
		return err
	}
	err = d.DB.Put([]byte(d.Key+state), []byte(rb), nil)
	if err != nil {
		dbError := fmt.Sprintf("%s can not put value", d.Key+state)
		logrus.Errorf(dbError)
		return errors.New(dbError)
	}
	return nil
}

func (d *Download) readDB() (error *DumpError, result *DumpResult, err error) {
	eb, err := d.DB.Get([]byte(d.Key+dError), nil)
	if err != nil {
		dbErr := fmt.Sprintf("%s can not get value", d.Key+dError)
		logrus.Errorf(dbErr)
		return nil, nil, errors.New(dbErr)
	}

	err = json.Unmarshal(eb, &error)
	if err != nil {
		jsonErr := fmt.Sprintf("Marshal %s dump error failed", d.Key)
		logrus.Errorf(jsonErr)
		return nil, nil, errors.New(jsonErr)
	}

	rb, err := d.DB.Get([]byte(d.Key+state), nil)
	if err != nil {
		dbErr := fmt.Sprintf("%s can not get value", d.Key+state)
		logrus.Errorf(dbErr)
		return nil, nil, errors.New(dbErr)
	}

	err = json.Unmarshal(rb, &result)
	if err != nil {
		jsonErr := fmt.Sprintf("Marshal %s dump error failed", d.Key)
		logrus.Errorf(jsonErr)
		return nil, nil, errors.New(jsonErr)
	}
	return error, result, nil
}

func (d *Download) javaDump() error {
	//cmd := exec.Command("docker", "-H", host, "exec", docker, "lss")
	var dumpError DumpError
	var result DumpResult

	if err := d.writeDB(&dumpError, &result); err != nil {
		return err
	}

	jDump := exec.Command("docker", "-H", d.Host, "exec", d.Docker, "jmap", "-dump:format=b,file="+d.App+dumpPostfix, "1")
	out, status := runCommand(jDump)
	if status != 0 {
		dumpError.Dump = out
		result.Dump = failed
		if err := d.writeDB(&dumpError, &result); err != nil {
			return err
		}
		return errors.New(out)
	}
	if strings.Contains(out, "allocate") {
		dumpError.Dump = "Cannot allocate memory"
		result.Copy = failed
		if err := d.writeDB(&dumpError, &result); err != nil {
			return err
		}
		return errors.New(out)
	}
	result.Dump = success
	if err := d.writeDB(&dumpError, &result); err != nil {
		return err
	}
	d.FilePath = "/" + d.App + dumpPostfix
	return nil
}

func (d *Download) FileCopy() error {
	logrus.Infof("start copy file %s", d.FilePath)
	copyError, result, err := d.readDB()
	if err != nil {
		return err
	}

	copy := exec.Command("docker", "-H", d.Host, "cp", d.Docker+":"+d.FilePath, tmp)
	//copy := exec.Command("docker", "-H", d.Host, "cp", d.Docker+":/test", "/tmp/trade-rpc.dump")
	out, status := runCommand(copy)
	if status != 0 {
		copyError.Copy = out
		result.Copy = failed
		if err := d.writeDB(copyError, result); err != nil {
			return err
		}
		return errors.New(out)
	}
	result.Copy = success
	if err := d.writeDB(copyError, result); err != nil {
		return err
	}
	return nil
}

func (d *Download) compressFile() error {
	fileName := GetFileName(d.FilePath)
	fmt.Println(fileName)
	logrus.Infof("start compress file %s", fileName)
	compressError, result, err := d.readDB()
	if err != nil {
		return err
	}

	compress := exec.Command("tar", "-zcvf", tmp+fileName+postfix, "-C", tmp, fileName)
	out, status := runCommand(compress)
	if status != 0 {
		compressError.Compress = out
		result.Compress = failed
		if err := d.writeDB(compressError, result); err != nil {
			return err
		}
		return errors.New(out)
	}
	result.Compress = success
	if err := d.writeDB(compressError, result); err != nil {
		return err
	}
	return nil
}

func (d *Download) cleanUp() {
	fileName := GetFileName(d.FilePath)
	logrus.Infof("start clean file %s", fileName)
	if d.Option == "dump" {
		cleanContainer := exec.Command("docker", "-H", d.Host, "exec", d.Docker, "rm", "-rf", d.FilePath)
		out1, status1 := runCommand(cleanContainer)
		if status1 != 0 {
			logrus.Error(out1)
		}
	}
	cleanHost := exec.Command("rm", "-rf", tmp+fileName, tmp+fileName+postfix)
	out2, status2 := runCommand(cleanHost)
	if status2 != 0 {
		logrus.Error(out2)
	}
}

func (d *Download) uploadToOSS() error {
	fileName := GetFileName(d.FilePath)
	logrus.Infof("start upload file %s", fileName)
	uploadError, result, err := d.readDB()
	if err != nil {
		return err
	}

	client, err := oss.New(endpoint, accessKeyId, accessKeySecret)
	if err != nil {
		uploadError.Upload = err.Error()
		result.Upload = failed
		if err2 := d.writeDB(uploadError, result); err != nil {
			return err2
		}
		return err
	}

	bucket, err := client.Bucket("terminus-dump")
	if err != nil {
		uploadError.Upload = err.Error()
		result.Upload = failed
		if err2 := d.writeDB(uploadError, result); err != nil {
			return err2
		}
		return err
	}

	err = bucket.PutObjectFromFile(d.User+"/"+fileName+postfix, tmp+fileName+postfix)
	if err != nil {
		uploadError.Upload = err.Error()
		result.Upload = failed
		if err2 := d.writeDB(uploadError, result); err != nil {
			return err2
		}
		return err
	}
	result.Upload = success
	if err := d.writeDB(uploadError, result); err != nil {
		return err
	}
	return nil
}

func (d *Download) runDump() {
	defer d.cleanUp()
	if err := d.javaDump(); err != nil {
		logrus.Error(err)
		d.cleanUp()
		return
	}
	if err := d.FileCopy(); err != nil {
		logrus.Error(err)
		d.cleanUp()
		return
	}
	if err := d.compressFile(); err != nil {
		logrus.Error(err)
		d.cleanUp()
		return
	}
	if err := d.uploadToOSS(); err != nil {
		logrus.Error(err)
		d.cleanUp()
		return
	}
	//javaDump("10.167.0.227", "8c0dcf0133c2", "trade-rpc")
	//dumpCopy("10.167.0.227", "8c0dcf0133c2", "trade-rpc")
	//compressDump("trade-rpc")
	//uploadToOSS("tangyujie", "trade-rpc")
	//cleanUp("10.167.0.227", "8c0dcf0133c2", "trade-rpc")
}

func (d *Download) runDownload() {
	defer d.cleanUp()

	switch d.Option {
	case "dump":
		if err := d.javaDump(); err != nil {
			logrus.Error(err)
			d.cleanUp()
			return
		}
	default:
		var dumpError DumpError
		var result DumpResult

		result.Dump = success
		if err := d.writeDB(&dumpError, &result); err != nil {
			logrus.Error("init db failed")
			return
		}
	}

	if err := d.FileCopy(); err != nil {
		logrus.Error(err)
		d.cleanUp()
		return
	}
	if err := d.compressFile(); err != nil {
		logrus.Error(err)
		d.cleanUp()
		return
	}
	if err := d.uploadToOSS(); err != nil {
		logrus.Error(err)
		d.cleanUp()
		return
	}
}

func DiceDownload(w http.ResponseWriter, r *http.Request) {
	var req GetDumpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		colonyutil.WriteErr(w, "400", err.Error())
		return
	}

	if !strings.Contains(req.FilePath, "/") || len(strings.Split(req.FilePath, "/")) < 2 {
		colonyutil.WriteErr(w, "403", "invalid file path")
		return
	}

	key := req.User + "-" + strconv.FormatInt(time.Now().Unix(), 10)

	go func() {
		download := &Download{
			GetDumpRequest: req,
			Key:            key,
			DB:             settings.DownloadDB,
		}
		download.runDownload()
	}()

	colonyutil.WriteData(w, &GetDumpResponse{
		Key:         key,
		Description: "accepted",
	})
	return
}

func ReadDownloadResult(w http.ResponseWriter, r *http.Request) {
	key := r.FormValue("key")
	download := &Download{
		Key: key,
		DB:  settings.DownloadDB,
	}
	uploadError, result, err := download.readDB()
	if err != nil {
		colonyutil.WriteErr(w, "500", err.Error())
		return
	}

	colonyutil.WriteData(w, &GetDumpStatusResponse{
		Data:  result,
		Error: uploadError,
	})
}

func GetFileName(path string) string {
	length := len(strings.Split(path, "/"))
	if strings.LastIndex(path, "/") == len(path)-1 {
		return strings.Split(path, "/")[length-2]
	}
	return strings.Split(path, "/")[length-1]
}
