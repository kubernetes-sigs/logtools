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

// Package gclplugin implements the golangci-lint's module plugin interface for logcheck to be used
// as a private linter in golangci-lint. See more details at
// https://golangci-lint.run/plugins/module-plugins/.
package gclplugin

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
	"sigs.k8s.io/logtools/logcheck/pkg"
)

func init() {
	register.Plugin("logcheck", New)
}

type settings struct {
	Check  map[string]bool `json:"check"`
	Config string          `json:"config"`
}

// New Module Plugin System, see https://golangci-lint.run/plugins/module-plugins/.
func New(pluginSettings interface{}) (register.LinterPlugin, error) {
	// We could manually parse the settings. This would involve several
	// type assertions. Encoding as JSON and then decoding into our
	// settings struct is easier.
	//
	// The downside is that format errors are less user-friendly.
	var buffer bytes.Buffer
	if err := json.NewEncoder(&buffer).Encode(pluginSettings); err != nil {
		return nil, fmt.Errorf("encoding settings as internal JSON buffer: %v", err)
	}
	var s settings
	decoder := json.NewDecoder(&buffer)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&s); err != nil {
		return nil, fmt.Errorf("decoding settings from internal JSON buffer: %v", err)
	}

	return &LogcheckPlugin{settings: s}, nil
}

var _ register.LinterPlugin = (*LogcheckPlugin)(nil)

type LogcheckPlugin struct {
	settings settings
}

// BuildAnalyzers implements register.LinterPlugin.
func (l *LogcheckPlugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	// Now create an analyzer and configure it.
	analyzer, config := pkg.Analyser()

	for check, enabled := range l.settings.Check {
		if err := config.SetEnabled(check, enabled); err != nil {
			// No need to wrap, the error is informative.
			return nil, err
		}
	}

	if err := config.ParseConfig(l.settings.Config); err != nil {
		return nil, fmt.Errorf("parsing config: %v", err)
	}

	return []*analysis.Analyzer{analyzer}, nil
}

// GetLoadMode implements register.LinterPlugin.
func (l *LogcheckPlugin) GetLoadMode() string { return register.LoadModeSyntax }
