/*
Copyright 2022 The Kubernetes Authors.

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

package pkg

import (
	"io/ioutil"
	"path"
	"testing"
)

func TestMatch(t *testing.T) {
	temp := t.TempDir()
	filename := path.Join(temp, "expressions")
	if err := ioutil.WriteFile(filename, []byte(`# Example file
structured hello
+structured a.c
-structured adc
structured x.*y
structured,parameters world
`), 0666); err != nil {
		t.Fatalf("writing file: %v", err)
	}

	filter := &RegexpFilter{
		validChecks: map[string]bool{
			structuredCheck: true,
			parametersCheck: true,
		},
	}
	if err := filter.Set(filename); err != nil {
		t.Fatalf("reading file: %v", err)
	}

	for _, tc := range []struct {
		filename      string
		check         string
		enabled       bool
		expectEnabled bool
	}{
		{
			filename:      "hello",
			check:         "structured",
			expectEnabled: true,
		},
		{
			filename:      "hello",
			check:         "parameters",
			expectEnabled: false, // not set
		},
		{
			filename:      "hello",
			check:         "parameters",
			enabled:       true,
			expectEnabled: true, // global default
		},
		{
			filename:      "hello/world",
			check:         "structured",
			expectEnabled: false, // no sub-matches
		},
		{
			filename:      "abc",
			check:         "structured",
			expectEnabled: true,
		},
		{
			filename:      "adc",
			check:         "structured",
			expectEnabled: false, // unset later
		},
		{
			filename:      "x1y",
			check:         "structured",
			expectEnabled: true,
		},
		{
			filename:      "x2y",
			check:         "structured",
			expectEnabled: true,
		},
	} {
		actualEnabled := filter.Enabled(tc.check, tc.enabled, tc.filename)
		if actualEnabled != tc.expectEnabled {
			t.Errorf("%+v: got %v", tc, actualEnabled)
		}
	}
}

func TestSetNoFile(t *testing.T) {
	filter := &RegexpFilter{}
	if err := filter.Set("no such file"); err == nil {
		t.Errorf("did not get expected error")
	}
}

func TestParsing(t *testing.T) {
	temp := t.TempDir()
	filename := path.Join(temp, "expressions")
	for name, tc := range map[string]struct {
		content     string
		expectError string
	}{
		"invalid-regexp": {
			content:     `structured [`,
			expectError: filename + ":0: error parsing regexp: missing closing ]: `[$`",
		},
		"invalid-line": {
			content: `structured .
parameters`,
			expectError: filename + ":1: not of the format <checks> <regexp>: parameters",
		},
		"invalid-check": {
			content:     `xxx .`,
			expectError: filename + ":0: \"xxx\" is not a supported check: xxx .",
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := ioutil.WriteFile(filename, []byte(tc.content), 0666); err != nil {
				t.Fatalf("writing file: %v", err)
			}

			filter := &RegexpFilter{
				validChecks: map[string]bool{
					structuredCheck: true,
					parametersCheck: true,
				},
			}
			err := filter.Set(filename)
			if tc.expectError == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("did not get expected error: %s", tc.expectError)
				}
				if err.Error() != tc.expectError {
					t.Fatalf("error mismatch\nexpected: %q\n     got: %q", tc.expectError, err.Error())
				}
			}
		})
	}
}
