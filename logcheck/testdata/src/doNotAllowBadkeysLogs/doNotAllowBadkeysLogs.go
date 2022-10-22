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

package doNotAllowBadkeysLogs

import (
	"github.com/go-logr/logr"
	klog "k8s.io/klog/v2"
)

var logger logr.Logger

func doNotAllowBadKeysLogs() {
	// Structured logs
	// Error is not expected as this package allows bad keys
	klog.InfoS("test log")
	klog.ErrorS(nil, "test log")
	klog.InfoS("Starting container in a pod", "containerID", "containerID", "pod")       // want `Additional arguments to InfoS should always be Key Value pairs. Please check if there is any key or value missing.`
	klog.ErrorS(nil, "Starting container in a pod", "containerID", "containerID", "pod") // want `Additional arguments to ErrorS should always be Key Value pairs. Please check if there is any key or value missing.`
	klog.InfoS("Starting container in a pod", "test", "containerID")
	klog.ErrorS(nil, "Starting container in a pod", "test", "containerID")
	klog.InfoS("Starting container in a pod", "TEST", "containerID")
	klog.ErrorS(nil, "Starting container in a pod", "TEST", "containerID")
	klog.InfoS("Starting container in a pod", "TESTs", "containerID")
	klog.ErrorS(nil, "Starting container in a pod", "TESTs", "containerID")
	klog.InfoS("Starting container in a pod", "测试", "containerID")              // want `Key positional arguments "测试" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	klog.ErrorS(nil, "Starting container in a pod", "测试", "containerID")        // want `Key positional arguments "测试" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	klog.InfoS("Starting container in a pod", " test", "containerID")           // want `Key positional arguments " test" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	klog.ErrorS(nil, "Starting container in a pod", " test", "containerID")     // want `Key positional arguments " test" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	klog.InfoS("Starting container in a pod", "test ", "containerID")           // want `Key positional arguments "test " are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	klog.ErrorS(nil, "Starting container in a pod", "test ", "containerID")     // want `Key positional arguments "test " are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	klog.InfoS("Starting container in a pod", "test test", "containerID")       // want `Key positional arguments "test test" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	klog.ErrorS(nil, "Starting container in a pod", "test test", "containerID") // want `Key positional arguments "test test" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	klog.InfoS("Starting container in a pod", "t*est", "containerID")           // want `Key positional arguments "t\*est" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	klog.ErrorS(nil, "Starting container in a pod", "t*est", "containerID")     // want `Key positional arguments "t\*est" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	klog.InfoS("Starting container in a pod", "test[0]", "containerID")         // want `Key positional arguments "test\[0\]" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	klog.ErrorS(nil, "Starting container in a pod", "test[0]", "containerID")   // want `Key positional arguments "test\[0\]" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	klog.InfoS("Starting container in a pod", "T", "containerID")               // want `Key positional arguments "T" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	klog.ErrorS(nil, "Starting container in a pod", "T", "containerID")         // want `Key positional arguments "T" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	klog.InfoS("Starting container in a pod", "Test", "containerID")            // want `Key positional arguments "Test" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	klog.ErrorS(nil, "Starting container in a pod", "Test", "containerID")      // want `Key positional arguments "Test" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	klog.InfoS("Starting container in a pod", "test1Test", "containerID")
	klog.ErrorS(nil, "Starting container in a pod", "test1Test", "containerID")
	klog.InfoS("Starting container in a pod", 7, "containerID")                                   // want `Key positional arguments are expected to be inlined constant strings. Please replace 7 provided with string value.`
	klog.ErrorS(nil, "Starting container in a pod", 7, "containerID")                             // want `Key positional arguments are expected to be inlined constant strings. Please replace 7 provided with string value.`
	klog.InfoS("Starting container in a pod", map[string]string{"test1": "value"}, "containerID") // want `Key positional arguments are expected to be inlined constant strings.`
	testKey := "a"
	klog.ErrorS(nil, "Starting container in a pod", testKey, "containerID") // want `Key positional arguments are expected to be inlined constant strings. Please replace testKey provided with string value.`

	// Error is not expected as this package allows unstructured bad keys
	klog.V(1).Infof("test log")
	klog.Infof("test log")
	klog.Info("test log")
	klog.Infoln("test log")
	klog.InfoDepth(1, "test log")
	klog.Warning("test log")
	klog.Warningf("test log")
	klog.WarningDepth(1, "test log")
	klog.Error("test log")
	klog.Errorf("test log")
	klog.Errorln("test log")
	klog.ErrorDepth(1, "test log")
	klog.Fatal("test log")
	klog.Fatalf("test log")
	klog.Fatalln("test log")
	klog.FatalDepth(1, "test log")
	klog.Exit("test log")
	klog.ExitDepth(1, "test log")
	klog.Exitln("test log")
	klog.Exitf("test log")

	logger.Info("test log")
	logger.Error(nil, "test log")
	logger.Info("Starting container in a pod", "containerID", "containerID", "pod")       // want `Additional arguments to Info should always be Key Value pairs. Please check if there is any key or value missing.`
	logger.Error(nil, "Starting container in a pod", "containerID", "containerID", "pod") // want `Additional arguments to Error should always be Key Value pairs. Please check if there is any key or value missing.`
	logger.WithValues("containerID", "containerID", "pod")                                // want `Additional arguments to WithValues should always be Key Value pairs. Please check if there is any key or value missing.`
	logger.Info("Starting container in a pod", "test", "containerID")
	logger.Error(nil, "Starting container in a pod", "test", "containerID")
	logger.WithValues("test", "containerID")
	logger.Info("Starting container in a pod", "TEST", "containerID")
	logger.Error(nil, "Starting container in a pod", "TEST", "containerID")
	logger.WithValues("TEST", "containerID")
	logger.Info("Starting container in a pod", "TESTs", "containerID")
	logger.Error(nil, "Starting container in a pod", "TESTs", "containerID")
	logger.WithValues("TESTs", "containerID")
	logger.Info("Starting container in a pod", "测试", "containerID")              // want `Key positional arguments "测试" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	logger.Error(nil, "Starting container in a pod", "测试", "containerID")        // want `Key positional arguments "测试" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	logger.WithValues("测试", "containerID")                                       // want `Key positional arguments "测试" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	logger.Info("Starting container in a pod", " test", "containerID")           // want `Key positional arguments " test" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	logger.Error(nil, "Starting container in a pod", " test", "containerID")     // want `Key positional arguments " test" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	logger.WithValues(" test", "containerID")                                    // want `Key positional arguments " test" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	logger.Info("Starting container in a pod", "test ", "containerID")           // want `Key positional arguments "test " are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	logger.Error(nil, "Starting container in a pod", "test ", "containerID")     // want `Key positional arguments "test " are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	logger.WithValues("test ", "containerID")                                    // want `Key positional arguments "test " are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	logger.Info("Starting container in a pod", "test test", "containerID")       // want `Key positional arguments "test test" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	logger.Error(nil, "Starting container in a pod", "test test", "containerID") // want `Key positional arguments "test test" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	logger.WithValues("test test", "containerID")                                // want `Key positional arguments "test test" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	logger.Info("Starting container in a pod", "t*est", "containerID")           // want `Key positional arguments "t\*est" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	logger.Error(nil, "Starting container in a pod", "t*est", "containerID")     // want `Key positional arguments "t\*est" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	logger.WithValues("t*est", "containerID")                                    // want `Key positional arguments "t\*est" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	logger.Info("Starting container in a pod", "test[0]", "containerID")         // want `Key positional arguments "test\[0\]" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	logger.Error(nil, "Starting container in a pod", "test[0]", "containerID")   // want `Key positional arguments "test\[0\]" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	logger.WithValues("test[0]", "containerID")                                  // want `Key positional arguments "test\[0\]" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	logger.Info("Starting container in a pod", "T", "containerID")               // want `Key positional arguments "T" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	logger.Error(nil, "Starting container in a pod", "T", "containerID")         // want `Key positional arguments "T" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	logger.WithValues("T", "containerID")                                        // want `Key positional arguments "T" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	logger.Info("Starting container in a pod", "Test", "containerID")            // want `Key positional arguments "Test" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	logger.Error(nil, "Starting container in a pod", "Test", "containerID")      // want `Key positional arguments "Test" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	logger.WithValues("Test", "containerID")                                     // want `Key positional arguments "Test" are expected to no special characters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.`
	logger.Info("Starting container in a pod", "test1Test", "containerID")
	logger.Error(nil, "Starting container in a pod", "test1Test", "containerID")
	logger.WithValues("test1Test", "containerID")
	logger.Info("Starting container in a pod", 7, "containerID")                                   // want `Key positional arguments are expected to be inlined constant strings. Please replace 7 provided with string value.`
	logger.Error(nil, "Starting container in a pod", 7, "containerID")                             // want `Key positional arguments are expected to be inlined constant strings. Please replace 7 provided with string value.`
	logger.WithValues(7, "containerID")                                                            // want `Key positional arguments are expected to be inlined constant strings. Please replace 7 provided with string value.`
	logger.Info("Starting container in a pod", map[string]string{"test1": "value"}, "containerID") // want `Key positional arguments are expected to be inlined constant strings.`
	logger.Error(nil, "Starting container in a pod", testKey, "containerID")                       // want `Key positional arguments are expected to be inlined constant strings. Please replace testKey provided with string value.`
	logger.WithValues(map[string]string{"test1": "value"}, "containerID")                          // want `Key positional arguments are expected to be inlined constant strings.`
}
