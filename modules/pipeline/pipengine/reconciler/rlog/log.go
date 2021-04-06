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

package rlog

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

const (
	errorFormat  = "reconciler: %s"
	pErrorFormat = "reconciler: pipelineID: %d, %s"
	tErrorFormat = "reconciler: pipelineID: %d, taskID: %d, %s"
)

func Infof(format string, args ...interface{}) {
	format = handleFormat(format)
	logrus.Infof(format, args...)
}

func PInfof(pipelineID uint64, format string, args ...interface{}) {
	format = handlePFormat(pipelineID, format)
	logrus.Infof(format, args...)
}

func TInfof(pipelineID, taskID uint64, format string, args ...interface{}) {
	format = handleTFormat(pipelineID, taskID, format)
	logrus.Infof(format, args...)
}

func PWarnf(pipelineID uint64, format string, args ...interface{}) {
	format = handlePFormat(pipelineID, format)
	logrus.Warnf(format, args...)
}

func TWarnf(pipelineID, taskID uint64, format string, args ...interface{}) {
	format = handleTFormat(pipelineID, taskID, format)
	format = handleAlert(format)
	logrus.Warnf(format, args...)
}

func PDebugf(pipelineID uint64, format string, args ...interface{}) {
	format = handlePFormat(pipelineID, format)
	logrus.Debugf(format, args...)
}

func TDebugf(pipelineID, taskID uint64, format string, args ...interface{}) {
	format = handleTFormat(pipelineID, taskID, format)
	logrus.Debugf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	format = handleFormat(format)
	format = handleAlert(format)
	logrus.Errorf(format, args...)
}

func PErrorf(pipelineID uint64, format string, args ...interface{}) {
	format = handlePFormat(pipelineID, format)
	format = handleAlert(format)
	logrus.Errorf(format, args...)
}

func TErrorf(pipelineID, taskID uint64, format string, args ...interface{}) {
	format = handleTFormat(pipelineID, taskID, format)
	format = handleAlert(format)
	logrus.Errorf(format, args...)
}

// ErrorAndReturn print polished log and return original error.
func ErrorAndReturn(err error) error {
	format := handleFormat(err.Error())
	format = handleAlert(format)
	logrus.Errorf(format)
	return err
}

// PErrorAndReturn print polished log and return original error.
func PErrorAndReturn(pipelineID uint64, err error) error {
	format := handlePFormat(pipelineID, err.Error())
	format = handleAlert(format)
	logrus.Errorf(format)
	return err
}

// TErrorAndReturn print polished log and return original error.
func TErrorAndReturn(pipelineID, taskID uint64, err error) error {
	format := handleTFormat(pipelineID, taskID, err.Error())
	format = handleAlert(format)
	logrus.Errorf(format)
	return err
}

func handleFormat(format string) string {
	return fmt.Sprintf(errorFormat, format)
}

func handlePFormat(pipelineID uint64, format string) string {
	return fmt.Sprintf(pErrorFormat, pipelineID, format)
}

func handleTFormat(pipelineID, taskID uint64, format string) string {
	return fmt.Sprintf(tErrorFormat, pipelineID, taskID, format)
}

func handleAlert(format string) string {
	return fmt.Sprintf("[alert] %s", format)
}
