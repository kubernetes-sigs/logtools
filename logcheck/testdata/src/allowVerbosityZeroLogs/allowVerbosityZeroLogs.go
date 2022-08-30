/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// This fake package is created as golang.org/x/tools/go/analysis/analysistest
// expects it to be here for loading. This package is used to test allow-unstructured
// flag which suppresses errors when unstructured logging is used.
// This is a test file for unstructured logging static check tool unit tests.

package Verbosity

import (
	"github.com/go-logr/logr"
	klog "k8s.io/klog/v2"
)

const (
	zeroConst = 0
	oneConst  = 1
)

var (
	zeroVar   klog.Level = 0
	oneVar    klog.Level = 1
	l, logger logr.Logger
)

func verbosityLogging() {
	klog.V(0).Info("test log")
	klog.V(0).Infof("test log")
	klog.V(0).Infoln("test log")
	klog.V(0).InfoS("I'm logging at level 0.")
	klog.V(zeroConst).InfoS("I'm logging at level 0.")
	klog.V(zeroVar).InfoS("I'm logging at level 0.")
	klog.V(1).Info("test log")
	klog.V(1).Infof("test log")
	klog.V(1).Infoln("test log")
	klog.V(1).InfoS("I'm logging at level 1.")
	klog.V(oneConst).InfoS("I'm logging at level 1.")
	klog.V(oneVar).InfoS("I'm logging at level 1.")
	klog.Info("test log")
	klog.Infof("test log")
	klog.Infoln("test log")
	klog.InfoS("I'm logging at level 0.")

	logger.Info("hello", "1", "2")
	logger.Error(nil, "hello", "1", "2")
	logger.WithValues("1", "2")

	logger.V(0).Info("hello", "1", "2")
	logger.V(0).Error(nil, "hello", "1", "2")
	logger.V(0).WithValues("1", "2")

	logger.V(1).Info("hello", "1", "2")
	logger.V(1).Error(nil, "hello", "1", "2")
	logger.V(1).WithValues("1", "2")
}
