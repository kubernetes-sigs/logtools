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

package contextual

import (
	"context"

	"github.com/go-logr/logr"
)

type myFuncType func(ctx context.Context, logger logr.Logger, msg string) // want `A function should accept either a context or a logger, but not both. Having both makes calling the function harder because it must be defined whether the context must contain the logger and callers have to follow that.`

func usingMyFuncType(firstParam int,
	callback myFuncType, // Will be warned about at the type definition, not here.
	lastParam int) {
}

func usingInlineFunc(firstParam int,
	callback func(ctx context.Context, logger logr.Logger, msg string), // want `A function should accept either a context or a logger, but not both. Having both makes calling the function harder because it must be defined whether the context must contain the logger and callers have to follow that.`
	lastParam int) {
}

type myStruct struct {
	myFuncField func(ctx context.Context, logger logr.Logger, msg string) // want `A function should accept either a context or a logger, but not both. Having both makes calling the function harder because it must be defined whether the context must contain the logger and callers have to follow that.`
}

func (m myStruct) myMethod(ctx context.Context, logger logr.Logger, msg string) { // want `A function should accept either a context or a logger, but not both. Having both makes calling the function harder because it must be defined whether the context must contain the logger and callers have to follow that.`
}

func myFunction(ctx context.Context, logger logr.Logger, msg string) { // want `A function should accept either a context or a logger, but not both. Having both makes calling the function harder because it must be defined whether the context must contain the logger and callers have to follow that.`
}

type myInterface interface {
	myInterfaceMethod(ctx context.Context, logger logr.Logger, msg string) // want `A function should accept either a context or a logger, but not both. Having both makes calling the function harder because it must be defined whether the context must contain the logger and callers have to follow that.`
}
