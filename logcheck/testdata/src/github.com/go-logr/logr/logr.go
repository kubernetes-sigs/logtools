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

// Package logr provides empty stubs for github.com/go-logr/logr for testing
// with golang.org/x/tools/go/analysis/analysistest.
package logr

type Logger struct{}

func (l Logger) Enabled() bool                                  { return false }
func (l Logger) WithName(name string) Logger                    { return l }
func (l Logger) WithValues(kv ...interface{}) Logger            { return l }
func (l Logger) V(level int) Logger                             { return l }
func (l Logger) Info(msg string, kv ...interface{})             {}
func (l Logger) Error(err error, msg string, kv ...interface{}) {}
