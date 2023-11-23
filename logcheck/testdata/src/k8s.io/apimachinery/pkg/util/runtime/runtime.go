/*
Copyright 2023 The Kubernetes Authors.

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

// This fake package is created as package golang.org/x/tools/go/analysis/analysistest
// expects test data dependency to be testdata/src
package runtime

import "context"

//logcheck:context // Use HandleErrorWithContext instead in code which supports contextual logging.
func HandleError(err error) { // want HandleError:"Use HandleErrorWithContext instead in code which supports contextual logging."
}

func HandleErrorWithContext(ctx context.Context, err error) {
}

//logcheck:xyz // want "unknown logcheck keyword in comment"
func Foo() {
}

//logcheck:context
func Bar() { // want Bar:"Bar should not be used in code which supports contextual logging."
}

type Handler interface {
	//logcheck:context//Use HandleErrorWithContext instead in code which supports contextual logging.
	HandleError(err error) // want HandleError:"Use HandleErrorWithContext instead in code which supports contextual logging."

	HandleErrorWithContext(ctx context.Context, err error)
}

func foo(ctx context.Context, err error) {
	HandleError(nil) // want "Use HandleErrorWithContext instead in code which supports contextual logging."
	HandleErrorWithContext(ctx, err)
	var h Handler
	h.HandleError(err) // want "Use HandleErrorWithContext instead in code which supports contextual logging."
	h.HandleErrorWithContext(ctx, err)
}
