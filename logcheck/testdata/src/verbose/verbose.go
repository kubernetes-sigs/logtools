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

package verbose

import (
	"github.com/go-logr/logr"
	klog "k8s.io/klog/v2"
)

var l, logger logr.Logger

func verboseLogging() {
	klog.V(1).Info("test log") // want `unstructured logging function "Info" should not be used`
	if klogV := klog.V(1); klogV.Enabled() {
		klogV.Infof("hello %s", "world") // want `unstructured logging function "Infof" should not be used`
	}

	// \(\) is actually () in the diagnostic output. We have to escape here
	// because `want` expects a regular expression.

	if klog.V(1).Enabled() { // want `the result of klog.V should be stored in a variable and then be used multiple times: if klogV := klog.V\(\); klogV.Enabled\(\) { ... klogV.Info ... }`
		klog.V(1).InfoS("I'm logging at level 1.")
	}

	if l.V(1).Enabled() { // want `the result of l.V should be stored in a variable and then be used multiple times: if l := l.V\(\); l.Enabled\(\) { ... l.Info ... }`
		l.V(1).Info("I'm logging at level 1.")
	}

	if l := l.V(2); l.Enabled() {
		l.Info("I'm logging at level 2.")
	}

	if l := logger.V(2); l.Enabled() {
		// This is probably an error (should be l instead of logger),
		// but not currently detected.
		logger.Info("I wanted to log at level 2, but really it is 0.")
	}
}
