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

// Package main is meant to be compiled as a plugin for golangci-lint, see
// https://golangci-lint.run/contributing/new-linters/#create-a-plugin.
package main

import (
	"github.com/mitchellh/mapstructure"
	"golang.org/x/tools/go/analysis"
	"sigs.k8s.io/logtools/logcheck/pkg"
)

type analyzerPlugin struct{}

func (*analyzerPlugin) GetAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		pkg.Analyser(),
	}
}

// AnalyzerPlugin is the entry point for golangci-lint.
var AnalyzerPlugin analyzerPlugin

// Settings is the value struct passed by golangci-lint.
type Settings struct {
	Config string
}

// New is the entry point for golangci-lint, which takes priority over AnalyzerPlugin.
func New(conf interface{}) ([]*analysis.Analyzer, error) {
	a := pkg.Analyser()

	settings := Settings{}
	if err := mapstructure.Decode(conf, &settings); err != nil {
		return nil, err
	}

	if settings.Config != "" {
		if err := a.Flags.Set("config", settings.Config); err != nil {
			return nil, err
		}
	}

	return []*analysis.Analyzer{a}, nil
}
