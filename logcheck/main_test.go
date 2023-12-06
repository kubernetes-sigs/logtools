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

package main

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"sigs.k8s.io/logtools/logcheck/pkg"
)

func TestAnalyzer(t *testing.T) {
	tests := []struct {
		name        string
		enabled     map[string]string
		override    string
		testPackage string
	}{
		{
			name: "Allow unstructured logs",
			enabled: map[string]string{
				"structured": "false",
				"key":        "false",
			},
			testPackage: "allowUnstructuredLogs",
		},
		{
			name: "Do not allow unstructured logs",
			enabled: map[string]string{
				"key": "false",
			},
			testPackage: "doNotAllowUnstructuredLogs",
		},
		{
			name: "Per-file config",
			enabled: map[string]string{
				"structured": "false",
			},
			override:    "testdata/src/mixed/structured_logging",
			testPackage: "mixed",
		},
		{
			name: "Function call parameters",
			enabled: map[string]string{
				"structured": "false",
				"key":        "false",
			},
			testPackage: "parameters",
		},
		{
			name: "Only allow contextual calls",
			enabled: map[string]string{
				"structured": "true",
				"contextual": "true",
			},
			testPackage: "onlyAllowContextual",
		},
		{
			name: "Only allow contextual calls through config",
			enabled: map[string]string{
				"structured": "false",
				"contextual": "false",
			},
			override:    "testdata/src/onlyAllowContextual/klog_logging",
			testPackage: "onlyAllowContextual",
		},
		{
			name:        "importrename",
			testPackage: "importrename",
		},
		{
			name:        "verbose",
			testPackage: "verbose",
		},
		{
			name: "gologr",
			enabled: map[string]string{
				"contextual": "true",
				"key":        "false",
			},
			testPackage: "gologr",
		},
		{
			name: "contextual",
			enabled: map[string]string{
				"contextual": "true",
			},
			testPackage: "contextual",
		},
		{
			name: "helpers",
			enabled: map[string]string{
				"with-helpers": "true",
			},
			testPackage: "helpers",
		},
		{
			name: "Do not allow Verbosity Zero logs",
			enabled: map[string]string{
				"structured": "false",
				"key":        "false",
			},
			testPackage: "doNotAllowVerbosityZeroLogs",
		},
		{
			name: "Allow Verbosity Zero logs",
			enabled: map[string]string{
				"structured":     "false",
				"verbosity-zero": "false",
				"key":            "false",
			},
			testPackage: "allowVerbosityZeroLogs",
		},
		{
			name: "Do not allow Bad Keys logs",
			enabled: map[string]string{
				"structured": "false",
				"parameters": "false",
			},
			testPackage: "doNotAllowBadkeysLogs",
		},
		{
			name: "Allow Bad Keys logs",
			enabled: map[string]string{
				"structured": "false",
				"parameters": "false",
				"key":        "false",
			},
			testPackage: "allowBadkeysLogs",
		},
		{
			name: "Detect incomplete fmt.Stringer",
			enabled: map[string]string{
				"value": "true",
			},
			testPackage: "stringer",
		},
		{
			name: "logcheck facts",
			enabled: map[string]string{
				"contextual": "true",
			},
			testPackage: "k8s.io/apimachinery/pkg/util/runtime",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			analyzer, _ := pkg.Analyser()
			set := func(flag, value string) {
				if value != "" {
					if err := analyzer.Flags.Set(flag, value); err != nil {
						t.Fatalf("unexpected error for %s: %v", flag, err)
					}
				}
			}
			for key, value := range tc.enabled {
				set("check-"+key, value)
			}
			if tc.override != "" {
				set("config", tc.override)
			}
			analysistest.Run(t, analysistest.TestData(), analyzer, tc.testPackage)
		})
	}
}
