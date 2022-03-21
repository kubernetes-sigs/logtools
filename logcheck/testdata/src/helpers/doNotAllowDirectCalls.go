/*
Copyright 2021 The Kubernetes Authors.

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

package helpers

import (
	"context"

	"github.com/go-logr/logr"
	klog "k8s.io/klog/v2"
)

var logger klog.Logger

func doNotAlllowDirectCalls() {
	logger.WithName("foo")                        // want `function "WithName" should be called through klogr.LoggerWithName`
	logger.WithValues("a", "b")                   // want `function "WithValues" should be called through klogr.LoggerWithValues`
	logr.NewContext(context.Background(), logger) // want `function "NewContext" should be called through klogr.NewContext`
}

func allowHelpers() {
	klog.LoggerWithName(logger, "foo")
	klog.LoggerWithValues(logger, "a", "b")
	klog.NewContext(context.Background(), logger)
}
