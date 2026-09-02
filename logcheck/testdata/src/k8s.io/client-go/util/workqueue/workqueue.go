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

// This fake package is created as package golang.org/x/tools/go/analysis/analysistest
// expects test data dependency to be testdata/src

package workqueue

import (
	klog "k8s.io/klog/v2"
)

type TypedInterface[T comparable] interface{}

// TypedDelayingQueueConfig specifies optional configurations to customize a
// TypedDelayingInterface, modeled after the real
// k8s.io/client-go/util/workqueue.TypedDelayingQueueConfig.
type TypedDelayingQueueConfig[T comparable] struct {
	// Logger is optional. If set, contextual logging is used, otherwise a
	// fallback logger is used.
	Logger *klog.Logger

	// Name for the queue.
	Name string
}

type DelayingQueueConfig = TypedDelayingQueueConfig[any]

func NewTypedDelayingQueueWithConfig[T comparable](config TypedDelayingQueueConfig[T]) TypedInterface[T] {
	return nil
}

func NewDelayingQueueWithConfig(config DelayingQueueConfig) TypedInterface[any] {
	return nil
}
