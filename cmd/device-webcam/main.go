// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

// This package provides a simple example of a device service.
package main

import (
	"github.com/redislabs/edgex-device-webcam"
	"github.com/redislabs/edgex-device-webcam/driver"
	"github.com/redislabs/edgex-device-webcam/pkg/startup"
)

const (
	serviceName string = "device-webcam"
)

func main() {
	wd := driver.WebcamDriver{}
	startup.Bootstrap(serviceName, device.Version, &wd)
}
