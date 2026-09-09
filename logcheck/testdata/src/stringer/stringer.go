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

package stringer

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	klog "k8s.io/klog/v2"
)

func foo() {
	klog.Background().Info("Starting", "config", config{})
	klog.Background().Info("Starting", "config", configWithStringer{})
	klog.Background().Info("Starting", "config", &config{}) // want `The type \*stringer.config inherits \(\*k8s.io/apimachinery/pkg/apis/meta/v1.TypeMeta\).String as implementation of fmt.Stringer, which covers only a subset of the value. Implement String\(\) for the type or wrap it with TODO.`
	klog.Background().Info("Starting", "config", &configWithStringer{})
	klog.Background().Info("Starting", "config", &simpleConfig{})
}

// config mimicks KubeletConfig (see
// https://github.com/kubernetes/kubernetes/pull/115950).  As far as logging is
// concerned, the type is broken: it implements fmt.Stringer because it
// embeds TypeMeta, but the result of String() is incomplete.
type config struct {
	metav1.TypeMeta // implements fmt.Stringer (but only for addressable values)

	RealField int
}

type configWithStringer config

func (c configWithStringer) Size() int {
	return 1
}

func (c configWithStringer) String() string {
	return "foo"
}

// simpleConfig only has a single field. In this case inheriting the String implementation
// is fine. This occurs for https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Time.
type simpleConfig struct {
	metav1.TypeMeta
}

func nilUnsafeStringers() {
	var valueStringer *timestamp
	var wrapperStringer *wrappedTimestamp
	var pointerStringer *nilSafeStringer
	var plainValue timestamp

	// Calling String() on a nil *timestamp panics because the method has a
	// value receiver and thus dereferences the pointer.
	klog.Background().Info("Starting", "time", valueStringer) // want `The type \*stringer.timestamp implements fmt.Stringer via the value receiver method \(stringer.timestamp\).String. Calling String\(\) panics for a nil pointer, which klog then logs instead of the value. Wrap the value with klog.SafePtr.`
	klog.InfoS("Starting", "time", valueStringer)             // want `The type \*stringer.timestamp implements fmt.Stringer via the value receiver method \(stringer.timestamp\).String. Calling String\(\) panics for a nil pointer, which klog then logs instead of the value. Wrap the value with klog.SafePtr.`

	klog.Background().Info("Starting", "time", wrapperStringer) // want `The type \*stringer.wrappedTimestamp implements fmt.Stringer via the value receiver method \(stringer.timestamp\).String. Calling String\(\) panics for a nil pointer, which klog then logs instead of the value. Wrap the value with klog.SafePtr.`

	// A pointer receiver can (and here does) handle nil itself.
	klog.Background().Info("Starting", "value", pointerStringer)

	// A non-pointer value cannot be nil.
	klog.Background().Info("Starting", "time", plainValue)

	// Taking the address of a value never yields a nil pointer.
	klog.Background().Info("Starting", "time", &plainValue)

	// SafePtr handles nil pointers, that is the suggested fix.
	klog.Background().Info("Starting", "time", klog.SafePtr(valueStringer))
}

// timestamp implements fmt.Stringer with a value receiver, like time.Time
// does. Calling String() through a nil pointer panics.
type timestamp struct {
	seconds int64
	nanos   int32
}

func (t timestamp) String() string {
	return "some point in time"
}

// wrappedTimestamp mimicks metav1.Time: a single-field wrapper struct that
// inherits the value receiver String() of the embedded type.
type wrappedTimestamp struct {
	timestamp
}

// nilSafeStringer implements fmt.Stringer with a pointer receiver that
// handles nil, like *net.IPNet does.
type nilSafeStringer struct {
	value string
}

func (s *nilSafeStringer) String() string {
	if s == nil {
		return "<nil>"
	}
	return s.value
}
