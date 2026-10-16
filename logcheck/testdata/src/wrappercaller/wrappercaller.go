/*
Copyright 2026 The Kubernetes Authors.

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

// This package tests cross-package wrapper detection.

package wrappercaller

import (
	"wrapperlib"
)

func callCrossPackageWrappers() {
	// Cross-package wrapper, odd args (should flag)
	wrapperlib.LogInfo("msg", "key") // want `Additional arguments to LogInfo should always be Key Value pairs. Please check if there is any key or value missing.`

	// Cross-package wrapper, correct args (should NOT flag)
	wrapperlib.LogInfo("msg", "key", "value")

	// Cross-package ErrorS wrapper, odd args (should flag)
	wrapperlib.LogError(nil, "msg", "key") // want `Additional arguments to LogError should always be Key Value pairs. Please check if there is any key or value missing.`

	// Cross-package wrapper, bad key (should flag)
	wrapperlib.LogInfo("msg", 1, "value") // want `Key positional arguments are expected to be inlined constant strings. Please replace 1 provided with string value`
}
