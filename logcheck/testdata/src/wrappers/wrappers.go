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

// This package tests the //logcheck:wrapper marker-based detection and
// linting of wrapper functions around structured logging calls.

package wrappers

import (
	"github.com/go-logr/logr"
	klog "k8s.io/klog/v2"
)

var logger logr.Logger

// --- Wrapper definitions (marked with //logcheck:wrapper) ---

//logcheck:wrapper
func logInfoKlog(msg string, kvs ...interface{}) { // want logInfoKlog:"logKVWrapper at index 1"
	klog.InfoS(msg, kvs...)
}

//logcheck:wrapper
func logErrorKlog(err error, msg string, kvs ...interface{}) { // want logErrorKlog:"logKVWrapper at index 2"
	klog.ErrorS(err, msg, kvs...)
}

//logcheck:wrapper
func logInfoLogr(l logr.Logger, msg string, kvs ...interface{}) { // want logInfoLogr:"logKVWrapper at index 2"
	l.Info(msg, kvs...)
}

//logcheck:wrapper
func logErrorLogr(l logr.Logger, err error, msg string, kvs ...interface{}) { // want logErrorLogr:"logKVWrapper at index 3"
	l.Error(err, msg, kvs...)
}

//logcheck:wrapper
func withValuesLogr(l logr.Logger, kvs ...interface{}) logr.Logger { // want withValuesLogr:"logKVWrapper at index 1"
	return l.WithValues(kvs...)
}

//logcheck:wrapper
func withValuesKlog(l logr.Logger, kvs ...interface{}) logr.Logger { // want withValuesKlog:"logKVWrapper at index 1"
	return klog.LoggerWithValues(l, kvs...)
}

// chainedWrapper wraps logInfoKlog. It is also marked as a wrapper so
// that its own call sites are checked.
//
//logcheck:wrapper
func chainedWrapper(msg string, kvs ...interface{}) { // want chainedWrapper:"logKVWrapper at index 1"
	logInfoKlog(msg, kvs...)
}

// Handler demonstrates a method wrapper.
type Handler struct{}

//logcheck:wrapper
func (h *Handler) Log(msg string, kvs ...interface{}) { // want Log:"logKVWrapper at index 1"
	klog.InfoS(msg, kvs...)
}

// notAWrapper is a variadic function without the marker, should NOT be
// treated as a wrapper.
func notAWrapper(msg string, args ...interface{}) {
	_ = msg
	_ = args
}

// --- Variable-based wrappers ---

// logViaAlias is a variable holding a direct reference to klog.InfoS.
// Auto-detected without a marker.
var logViaAlias = klog.InfoS // want logViaAlias:"logKVWrapper at index 1"

// logViaErrorAlias is a variable holding a direct reference to klog.ErrorS.
// Auto-detected without a marker.
var logViaErrorAlias = klog.ErrorS // want logViaErrorAlias:"logKVWrapper at index 2"

// logViaFuncVar is a function variable with the marker but unknown implementation.
//
//logcheck:wrapper
var logViaFuncVar func(msg string, kvs ...interface{}) // want logViaFuncVar:"logKVWrapper at index 1"

// --- Call sites ---

func callSites() {
	// Basic klog.InfoS wrapper, odd args (should flag)
	logInfoKlog("msg", "key") // want `Additional arguments to logInfoKlog should always be Key Value pairs. Please check if there is any key or value missing.`

	// Basic klog.InfoS wrapper, bad key (should flag)
	logInfoKlog("msg", 1, "value") // want `Key positional arguments are expected to be inlined constant strings. Please replace 1 provided with string value`

	// Basic klog.InfoS wrapper, correct args (should NOT flag)
	logInfoKlog("msg", "key", "value")

	// klog.ErrorS wrapper, odd args (should flag)
	logErrorKlog(nil, "msg", "key") // want `Additional arguments to logErrorKlog should always be Key Value pairs. Please check if there is any key or value missing.`

	// klog.ErrorS wrapper, correct args (should NOT flag)
	logErrorKlog(nil, "msg", "key", "value")

	// logr.Logger.Info wrapper, odd args (should flag)
	logInfoLogr(logger, "msg", "key") // want `Additional arguments to logInfoLogr should always be Key Value pairs. Please check if there is any key or value missing.`

	// logr.Logger.Error wrapper, odd args (should flag)
	logErrorLogr(logger, nil, "msg", "key") // want `Additional arguments to logErrorLogr should always be Key Value pairs. Please check if there is any key or value missing.`

	// logr.Logger.WithValues wrapper, odd args (should flag)
	withValuesLogr(logger, "key") // want `Additional arguments to withValuesLogr should always be Key Value pairs. Please check if there is any key or value missing.`

	// klog.LoggerWithValues wrapper, odd args (should flag)
	withValuesKlog(logger, "key") // want `Additional arguments to withValuesKlog should always be Key Value pairs. Please check if there is any key or value missing.`

	// Chained wrapper, odd args (should flag)
	chainedWrapper("msg", "key") // want `Additional arguments to chainedWrapper should always be Key Value pairs. Please check if there is any key or value missing.`

	// Method wrapper, odd args (should flag)
	h := &Handler{}
	h.Log("msg", "key") // want `Additional arguments to Log should always be Key Value pairs. Please check if there is any key or value missing.`

	// Wrapper called with ellipsis, should NOT flag
	kvs := []interface{}{"key1", "value1"}
	logInfoKlog("msg", kvs...)

	// Not a wrapper, should NOT flag
	notAWrapper("msg", "key")

	// Variable alias wrapper (auto-detected), odd args (should flag)
	logViaAlias("msg", "key") // want `Additional arguments to logViaAlias should always be Key Value pairs. Please check if there is any key or value missing.`

	// Variable alias wrapper (auto-detected), correct args (should NOT flag)
	logViaAlias("msg", "key", "value")

	// ErrorS alias (auto-detected), odd args (should flag)
	logViaErrorAlias(nil, "msg", "key") // want `Additional arguments to logViaErrorAlias should always be Key Value pairs. Please check if there is any key or value missing.`

	// ErrorS alias (auto-detected), correct args (should NOT flag)
	logViaErrorAlias(nil, "msg", "key", "value")

	// Function variable wrapper, odd args (should flag)
	logViaFuncVar("msg", "key") // want `Additional arguments to logViaFuncVar should always be Key Value pairs. Please check if there is any key or value missing.`

	// Function variable wrapper, correct args (should NOT flag)
	logViaFuncVar("msg", "key", "value")
}
