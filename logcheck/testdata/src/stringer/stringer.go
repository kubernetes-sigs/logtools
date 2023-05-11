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
