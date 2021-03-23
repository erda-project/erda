package actionagent

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/actionagent/agenttool"
	"github.com/erda-project/erda/pkg/filehelper"
	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/erda-project/erda/pkg/retry"
)

const (
	logUploadFilePrefix = "[upload files] "
)

func (agent *Agent) uploadDir() {
	// uploadDir 未指定，不上传
	if agent.EasyUse.ContainerUploadDir == "" {
		return
	}

	// 校验目录是否存在
	if err := filehelper.CheckExist(agent.EasyUse.ContainerUploadDir, true); err != nil {
		agent.AppendError(err)
		return
	}

	// 遍历目录
	files, err := ioutil.ReadDir(agent.EasyUse.ContainerUploadDir)
	if err != nil {
		agent.AppendError(err)
		return
	}
	if len(files) == 0 {
		return
	}

	logrus.Println(logUploadFilePrefix + "begin")
	defer logrus.Println(logUploadFilePrefix + "done")

	var needUploadFile []*os.File
	for _, f := range files {
		var (
			file *os.File
			err  error
		)
		// 目录压缩上传
		if f.IsDir() {
			tarName := filepath.Join(agent.EasyUse.ContainerTempTarUploadDir, f.Name()+".tar")
			err := agenttool.Tar(tarName, filepath.Join(agent.EasyUse.ContainerUploadDir, f.Name()))
			if err != nil {
				agent.AppendError(err)
				continue
			}
			file, err = os.Open(tarName)
			if err != nil {
				agent.AppendError(err)
				continue
			}
		} else {
			// 普通文件直接上传
			file, err = os.Open(filepath.Join(agent.EasyUse.ContainerUploadDir, f.Name()))
			if err != nil {
				agent.AppendError(err)
				continue
			}
		}
		needUploadFile = append(needUploadFile, file)
	}
	// 文件上传限制
	// 1. 单个文件大小 10MB
	// 2. 总大小 50MB
	const (
		singleFileSizeLimit = datasize.MB * 10
		totalFileSizeLimit  = datasize.MB * 50
	)
	var currentTotalFileSize datasize.ByteSize
	var uploadedFiles []*apistructs.File
	for _, f := range needUploadFile {
		// 判断单个文件大小
		fileInfo, err := f.Stat()
		if err != nil {
			agent.AppendError(err)
			continue
		}
		currentFileSize := datasize.ByteSize(fileInfo.Size())
		if currentFileSize > singleFileSizeLimit {
			logrus.Printf(logUploadFilePrefix+"ignore file to upload, file is too large, fileName: %s, size: %s, singleFileLimit: %s\n",
				fileInfo.Name(), currentFileSize.HumanReadable(), singleFileSizeLimit.HumanReadable())
			continue
		}
		if currentTotalFileSize+currentFileSize > totalFileSizeLimit {
			logrus.Printf(logUploadFilePrefix+"ignore file to upload, total file size is too large, fileName: %s, fileSize: %s, currentTotalSize: %s, totalLimit: %s\n",
				fileInfo.Name(), currentFileSize.HumanReadable(), currentTotalFileSize.HumanReadable(), totalFileSizeLimit.HumanReadable())
			continue
		}
		// 上传
		diceFile, err := agent.uploadFile(f)
		if err != nil {
			logrus.Printf(logUploadFilePrefix+"upload failed, fileName: %s, size: %s, err: %v\n", fileInfo.Name(), currentFileSize.HumanReadable(), err)
			continue
		}
		currentTotalFileSize += currentFileSize
		logrus.Printf(logUploadFilePrefix+"upload success, fileName: %s, size: %s, currentTotalSize: %s", fileInfo.Name(), currentFileSize.HumanReadable(), currentTotalFileSize.HumanReadable())
		uploadedFiles = append(uploadedFiles, diceFile)
	}
	// put into metafile
	var metadata apistructs.Metadata
	for _, f := range uploadedFiles {
		metadata = append(metadata, apistructs.MetadataField{
			Name:  f.DisplayName,
			Value: f.UUID,
			Type:  apistructs.MetadataTypeDiceFile,
		})
	}
	err = agent.callbackToPipelinePlatform(&callback{Metadata: metadata})
	if err != nil {
		agent.AppendError(err)
	}
}

func (agent *Agent) uploadFile(file *os.File) (*apistructs.File, error) {
	var uploadResp apistructs.FileUploadResponse

	err := retry.DoWithInterval(func() error {
		resp, err := httpclient.New(httpclient.WithCompleteRedirect()).
			Post(agent.EasyUse.OpenAPIAddr).
			Path("/api/files").
			Param("fileFrom", fmt.Sprintf("action-upload-%d-%d", agent.Arg.PipelineID, agent.Arg.PipelineTaskID)).
			Param("expiredIn", "168h").
			Header("Authorization", agent.EasyUse.TokenForBootstrap).
			MultipartFormDataBody(map[string]httpclient.MultipartItem{
				"file": {Reader: file},
			}).
			Do().
			JSON(&uploadResp)
		if err != nil {
			return err
		}
		if !resp.IsOK() || !uploadResp.Success {
			return fmt.Errorf("statusCode: %d, respError: %s", resp.StatusCode(), uploadResp.Error)
		}
		return nil
	}, 5, time.Second*5)
	if err != nil {
		return nil, err
	}

	return uploadResp.Data, nil
}
