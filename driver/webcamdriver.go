// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 Canonical Ltd
// Copyright (C) 2018-2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

// This package provides a simple example implementation of
// ProtocolDriver interface.
//
package driver

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	dsModels "github.com/redislabs/edgex-device-webcam/pkg/models"
	redisedge "github.com/redislabs/edgex-device-webcam/redisedge"
)

// WebcamDriver represents a webcam
type WebcamDriver struct {
	lc      logger.LoggingClient
	asyncCh chan<- *dsModels.AsyncValues
	re      *redisedge.RedisEdge
}

func getImageBytes(imgFile string, buf *bytes.Buffer) error {
	// Read existing image from file
	img, err := os.Open(imgFile)
	if err != nil {
		return err
	}
	defer img.Close()

	// TODO: Attach MediaType property, determine if decoding
	//  early is required (to optimize edge processing)

	// Expect "png" or "jpeg" image type
	imageData, imageType, err := image.Decode(img)
	if err != nil {
		return err
	}
	// Finished with file. Reset file pointer
	img.Seek(0, 0)
	if imageType == "jpeg" {
		err = jpeg.Encode(buf, imageData, nil)
		if err != nil {
			return err
		}
	} else if imageType == "png" {
		err = png.Encode(buf, imageData)
		if err != nil {
			return err
		}
	}
	return nil
}

// Initialize performs protocol-specific initialization for the device
// service.
func (w *WebcamDriver) Initialize(lc logger.LoggingClient, asyncCh chan<- *dsModels.AsyncValues) error {
	w.lc = lc
	w.asyncCh = asyncCh
	w.lc.Debug("WebcamDriver.Initialize called")

	// TODO: Device Configuration should be begotten from Config.Device
	dc := map[string]string{
		"RedisURL":        "redis://localhost:6379",
		"RAIModelKey":     "redisai:model:yolo",
		"RAIModelPath":    "res/models/tiny-yolo-voc.pb",
		"RAIModelBackend": "TF",
		"RAIModelDevice":  "CPU",
		"RAIScriptKey":    "redisai:script:yolo-boxes",
		"RAIScriptPath":   "res/scripts/yolo-boxes.py",
		"RAIScriptDevice": "CPU",
	}
	re, err := redisedge.Initialize(dc, lc)
	if err != nil {
		w.lc.Error(fmt.Sprintf("WebcamDriver.Initialize: Error while initializing RedisEdge - %v", err))
		return err
	}
	w.re = re
	w.lc.Debug("WebcamDriver.Initialize exited")
	return nil
}

// HandleReadCommands triggers a protocol Read operation for the specified device.
func (w *WebcamDriver) HandleReadCommands(deviceName string, protocols map[string]contract.ProtocolProperties, reqs []dsModels.CommandRequest) (res []*dsModels.CommandValue, err error) {
	w.lc.Debug("WebcamDriver.HandleReadCommand called")

	if len(reqs) != 1 && len(reqs) != 3 {
		err = fmt.Errorf("WebcamDriver.HandleReadCommands; either 1 or 3 command requests are supported")
		return
	}
	w.lc.Debug(fmt.Sprintf("WebcamDriver.HandleReadCommands: protocols: %v resource: %v attributes: %v", protocols, reqs[0].DeviceResourceName, reqs[0].Attributes))

	res = make([]*dsModels.CommandValue, len(reqs))

	// The first request should always be for a frame
	if reqs[0].DeviceResourceName != "Frame" {
		err = fmt.Errorf("WebcamDriver.HandleReadCommands; Frame must be the first request")
		return
	}
	now := time.Now().UnixNano() / int64(time.Millisecond)
	buf := new(bytes.Buffer)
	// TODO: get an actual video frame
	err = getImageBytes("./res/sample_dog_416.jpg", buf)
	cvb, _ := dsModels.NewBinaryValue(reqs[0].DeviceResourceName, now, buf.Bytes())
	res[0] = cvb
	w.lc.Debug(fmt.Sprintf("WebcamDriver.HandleReadCommands: read frame %v", res[0]))

	// Get detections
	if len(reqs) > 1 {
		hoomans, doggos, e := w.re.YOLODetect(buf.Bytes())
		if e != nil {
			err = e
			w.lc.Debug(fmt.Sprintf("WebcamDriver.HandleReadCommands: YOLODetect failed - %v", err))
			return 
		}
		cv, _ := dsModels.NewUint64Value(reqs[1].DeviceResourceName, now, hoomans)
		res[1] = cv
		cv, _ = dsModels.NewUint64Value(reqs[2].DeviceResourceName, now, doggos)
		res[2] = cv
	}
	w.lc.Debug("WebcamDriver.HandleReadCommand existed")

	return
}

// HandleWriteCommands passes a slice of CommandRequest struct each representing
// a ResourceOperation for a specific device resource.
// Since the commands are actuation commands, params provide parameters for the individual
// command.
func (w *WebcamDriver) HandleWriteCommands(deviceName string, protocols map[string]contract.ProtocolProperties, reqs []dsModels.CommandRequest,
	params []*dsModels.CommandValue) error {
	w.lc.Debug("WebcamDriver.HandleWriteCommands called")

	if len(reqs) != 1 {
		err := fmt.Errorf("WebcamDriver.HandleWriteCommands; too many command requests; only one supported")
		return err
	}
	if len(params) != 1 {
		err := fmt.Errorf("WebcamDriver.HandleWriteCommands; the number of parameter is not correct; only one supported")
		return err
	}

	w.lc.Debug(fmt.Sprintf("WebcamDriver.HandleWriteCommands: protocols: %v, resource: %v, parameters: %v", protocols, reqs[0].DeviceResourceName, params))
	// TODO: make the FPS and input source configurable
	// var err error
	// if s.switchButton, err = params[0].BoolValue(); err != nil {
	// 	err := fmt.Errorf("WebcamDriver.HandleWriteCommands; the data type of parameter should be Boolean, parameter: %s", params[0].String())
	// 	return err
	// }
	w.lc.Debug("WebcamDriver.HandleWriteCommands exited")
	return nil
}

// Stop the protocol-specific DS code to shutdown gracefully, or
// if the force parameter is 'true', immediately. The driver is responsible
// for closing any in-use channels, including the channel used to send async
// readings (if supported).
func (w *WebcamDriver) Stop(force bool) error {
	w.lc.Debug(fmt.Sprintf("WebcamDriver.Stop called: force=%v", force))
	return nil
}
