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

package onlyallowcontextual

import (
	"context"

	"k8s.io/apimachinery/pkg/util/runtime"
	klog "k8s.io/klog/v2"
)

func doNotAlllowKlog() {
	klog.InfoS("test log")       // want `function "InfoS" should not be used, convert to contextual logging`
	klog.ErrorS(nil, "test log") // want `function "ErrorS" should not be used, convert to contextual logging`
	klog.V(1).Infof("test log")  // want `function "V" should not be used, convert to contextual logging` `function "Infof" should not be used, convert to contextual logging`

	klog.KObjs(nil)                                // want `Detected usage of deprecated helper "KObjs". Please switch to "KObjSlice" instead.`
	klog.InfoS("test log", "key", klog.KObjs(nil)) // want `function "InfoS" should not be used, convert to contextual logging` `Detected usage of deprecated helper "KObjs". Please switch to "KObjSlice" instead.`
}

func doNotAllNonContext(ctx context.Context) {
	runtime.HandleError(nil) // want `Use HandleErrorWithContext instead in code which supports contextual logging.`
	runtime.HandleErrorWithContext(ctx, nil)
	var handler runtime.Handler
	handler.HandleError(nil) // want `Use HandleErrorWithContext instead in code which supports contextual logging.`
	handler.HandleErrorWithContext(ctx, nil)
}
