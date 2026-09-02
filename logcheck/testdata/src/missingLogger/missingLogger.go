/*
Copyright 2024 The Kubernetes Authors.

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

package missingLogger

import (
	"k8s.io/client-go/util/workqueue"
	klog "k8s.io/klog/v2"
)

// direct composite literal, Logger field missing entirely.
func directLiteralMissing() {
	workqueue.NewTypedDelayingQueueWithConfig(workqueue.TypedDelayingQueueConfig[string]{ // want `the Logger field of k8s.io/client-go/util/workqueue.TypedDelayingQueueConfig\[string\] is not set, contextual logging will not work as expected unless a fallback logger is acceptable here`
		Name: "foo",
	})
}

// direct composite literal, Logger field set explicitly to nil.
func directLiteralNil() {
	workqueue.NewTypedDelayingQueueWithConfig(workqueue.TypedDelayingQueueConfig[string]{ // want `the Logger field of k8s.io/client-go/util/workqueue.TypedDelayingQueueConfig\[string\] is not set, contextual logging will not work as expected unless a fallback logger is acceptable here`
		Logger: nil,
		Name:   "foo",
	})
}

// direct composite literal, Logger field set properly.
func directLiteralSet(logger klog.Logger) {
	workqueue.NewTypedDelayingQueueWithConfig(workqueue.TypedDelayingQueueConfig[string]{
		Logger: &logger,
		Name:   "foo",
	})
}

// local variable initialized with a composite literal, Logger field missing,
// no later assignment either.
func localVariableMissing() {
	config := workqueue.TypedDelayingQueueConfig[string]{
		Name: "foo",
	}
	workqueue.NewTypedDelayingQueueWithConfig(config) // want `the Logger field of k8s.io/client-go/util/workqueue.TypedDelayingQueueConfig\[string\] is not set, contextual logging will not work as expected unless a fallback logger is acceptable here`
}

// local variable initialized with a composite literal, Logger field set
// afterwards through a field assignment.
func localVariableSetLater(logger klog.Logger) {
	config := workqueue.TypedDelayingQueueConfig[string]{
		Name: "foo",
	}
	config.Logger = &logger
	workqueue.NewTypedDelayingQueueWithConfig(config)
}

// local variable initialized with a composite literal, Logger field set and
// then reset to nil again.
func localVariableUnsetLater(logger klog.Logger) {
	config := workqueue.TypedDelayingQueueConfig[string]{
		Logger: &logger,
		Name:   "foo",
	}
	config.Logger = nil
	workqueue.NewTypedDelayingQueueWithConfig(config) // want `the Logger field of k8s.io/client-go/util/workqueue.TypedDelayingQueueConfig\[string\] is not set, contextual logging will not work as expected unless a fallback logger is acceptable here`
}

// local variable initialized with a composite literal, Logger field set
// afterwards through a field assignment, but only inside an if statement.
// This must not warn: the caller does support contextual logging, just not
// unconditionally.
func localVariableSetInIf(useLogger bool, logger klog.Logger) {
	config := workqueue.TypedDelayingQueueConfig[string]{
		Name: "foo",
	}
	if useLogger {
		config.Logger = &logger
	}
	workqueue.NewTypedDelayingQueueWithConfig(config)
}

// The non-generic type alias also gets detected.
func directLiteralMissingAlias() {
	workqueue.NewDelayingQueueWithConfig(workqueue.DelayingQueueConfig{ // want `the Logger field of k8s.io/client-go/util/workqueue.DelayingQueueConfig is not set, contextual logging will not work as expected unless a fallback logger is acceptable here`
		Name: "foo",
	})
}

// A function which merely forwards a configuration struct that it received
// as a parameter cannot be analyzed locally and thus must not be flagged.
func forwardConfig(config workqueue.TypedDelayingQueueConfig[string]) {
	workqueue.NewTypedDelayingQueueWithConfig(config)
}

// A configuration struct obtained from some other function call cannot be
// analyzed locally either.
func fromOtherCall() {
	workqueue.NewTypedDelayingQueueWithConfig(makeConfig())
}

func makeConfig() workqueue.TypedDelayingQueueConfig[string] {
	return workqueue.TypedDelayingQueueConfig[string]{}
}
