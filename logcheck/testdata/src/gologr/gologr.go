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

package gologr

import (
	"github.com/go-logr/logr"
)

var logger logr.Logger

func logging() {
	logger.Info("hello", "missing value") // want `Additional arguments to Info should always be Key Value pairs. Please check if there is any key or value missing.`
	logger.Error(nil, "hello", 1, 2)      // want `Key positional arguments are expected to be inlined constant strings. Please replace 1 provided with string value`
	logger.WithValues("missing value")    // want `Additional arguments to WithValues should always be Key Value pairs. Please check if there is any key or value missing.`

	logger.V(1).Info("hello", "missing value") // want `Additional arguments to Info should always be Key Value pairs. Please check if there is any key or value missing.`
	logger.V(1).Error(nil, "hello", 1, 2)      // want `Key positional arguments are expected to be inlined constant strings. Please replace 1 provided with string value`
	logger.V(1).WithValues("missing value")    // want `Additional arguments to WithValues should always be Key Value pairs. Please check if there is any key or value missing.`

	// variadic input to logger.Info, logger.Error, logger.WithValues functions
	kvs := []interface{}{"key1", "value1"}
	logger.Info("foo message", kvs...)
	logger.Error(nil, "foo error message", kvs...)
	logger.WithValues(kvs...)
	logger.WithValues(kvs)                      // want `Additional arguments to WithValues should always be Key Value pairs. Please check if there is any key or value missing.`
	logger.Error(nil, "foo error message", kvs) // want `Additional arguments to Error should always be Key Value pairs. Please check if there is any key or value missing.`
}
