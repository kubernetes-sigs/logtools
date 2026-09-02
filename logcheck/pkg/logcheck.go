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

package pkg

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/exp/utf8string"
	"golang.org/x/tools/go/analysis"
)

const (
	structuredCheck     = "structured"
	parametersCheck     = "parameters"
	contextualCheck     = "contextual"
	withHelpersCheck    = "with-helpers"
	verbosityZeroCheck  = "verbosity-zero"
	verbosityErrorCheck = "verbosity-error"
	keyCheck            = "key"
	valueCheck          = "value"
	deprecationsCheck   = "deprecations"
	loggerCheck         = "logger-constructor"
)

type checks map[string]*bool

type Config struct {
	enabled       checks
	fileOverrides RegexpFilter
}

func (c Config) isEnabled(check string, filename string) bool {
	return c.fileOverrides.Enabled(check, *c.enabled[check], filename)
}

func (c *Config) SetEnabled(check string, enabled bool) error {
	_, ok := c.enabled[check]
	if !ok {
		return fmt.Errorf("unsupported check %q", check)
	}
	*c.enabled[check] = enabled
	return nil
}

func (c *Config) ParseConfig(configContent string) error {
	return c.fileOverrides.Parse(bytes.NewBufferString(configContent), "<buffer>")
}

// Analyser creates a new logcheck analyser.
func Analyser() (*analysis.Analyzer, *Config) {
	c := Config{
		enabled: checks{
			structuredCheck:     new(bool),
			parametersCheck:     new(bool),
			contextualCheck:     new(bool),
			withHelpersCheck:    new(bool),
			verbosityZeroCheck:  new(bool),
			verbosityErrorCheck: new(bool),
			keyCheck:            new(bool),
			valueCheck:          new(bool),
			deprecationsCheck:   new(bool),
			loggerCheck:         new(bool),
		},
	}
	c.fileOverrides.validChecks = map[string]bool{}
	for key := range c.enabled {
		c.fileOverrides.validChecks[key] = true
	}
	var logcheckFlags flag.FlagSet
	prefix := "check-"
	logcheckFlags.BoolVar(c.enabled[structuredCheck], prefix+structuredCheck, true, `When true, logcheck will warn about calls to unstructured
klog methods (Info, Infof, Error, Errorf, Warningf, etc).`)
	logcheckFlags.BoolVar(c.enabled[parametersCheck], prefix+parametersCheck, true, `When true, logcheck will check parameters of structured logging calls.`)
	logcheckFlags.BoolVar(c.enabled[contextualCheck], prefix+contextualCheck, false, `When true, logcheck will only allow log calls for contextual logging (retrieving a Logger from klog or the context and logging through that) and warn about all others.`)
	logcheckFlags.BoolVar(c.enabled[withHelpersCheck], prefix+withHelpersCheck, false, `When true, logcheck will warn about direct calls to WithName, WithValues and NewContext.`)
	logcheckFlags.BoolVar(c.enabled[verbosityZeroCheck], prefix+verbosityZeroCheck, true, `When true, logcheck will check whether the parameter for V() is 0.`)
	logcheckFlags.BoolVar(c.enabled[verbosityErrorCheck], prefix+verbosityErrorCheck, true, `When true, logcheck will check for V() in front of a logr Error call.`)
	logcheckFlags.BoolVar(c.enabled[keyCheck], prefix+keyCheck, true, `When true, logcheck will check whether name arguments are valid keys according to the guidelines in (https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments).`)
	logcheckFlags.BoolVar(c.enabled[valueCheck], prefix+valueCheck, false, `When true, logcheck will check for problematic values (for example, types that have an incomplete fmt.Stringer implementation).`)
	logcheckFlags.BoolVar(c.enabled[deprecationsCheck], prefix+deprecationsCheck, true, `When true, logcheck will analyze the usage of deprecated Klog function calls.`)
	logcheckFlags.BoolVar(c.enabled[loggerCheck], prefix+loggerCheck, false, `When true and the "contextual" check is also enabled, logcheck will warn about call sites which construct a configuration struct with an optional "Logger *logr.Logger" (or klog.Logger) field without setting that field, based on a best-effort data flow analysis. Call sites where the configuration struct cannot be analyzed (for example, because it was received as a parameter) are silently skipped.`)
	logcheckFlags.Var(&c.fileOverrides, "config", `A file which overrides the global settings for checks on a per-file basis via regular expressions.`)

	// Use env variables as defaults. This is necessary when used as plugin
	// for golangci-lint because of
	// https://github.com/golangci/golangci-lint/issues/1512.
	for key, enabled := range c.enabled {
		envVarName := "LOGCHECK_" + strings.ToUpper(strings.ReplaceAll(string(key), "-", "_"))
		if value, ok := os.LookupEnv(envVarName); ok {
			v, err := strconv.ParseBool(value)
			if err != nil {
				panic(fmt.Errorf("%s=%q: %v", envVarName, value, err))
			}
			*enabled = v
		}
	}
	if value, ok := os.LookupEnv("LOGCHECK_CONFIG"); ok {
		if err := c.fileOverrides.Set(value); err != nil {
			panic(fmt.Errorf("LOGCHECK_CONFIG=%q: %v", value, err))
		}
	}

	return &analysis.Analyzer{
		Name: "logcheck",
		Doc:  "Tool to check logging calls.",
		Run: func(pass *analysis.Pass) (interface{}, error) {
			return run(pass, &c)
		},
		Flags:     logcheckFlags,
		FactTypes: []analysis.Fact{new(warnContextual)},
	}, &c
}

// warnContextual is a fact that is set for methods or functions which have the
// `logcheck:context // <comment>` comment. The value stored here is that
// comment.
type warnContextual string

func (w warnContextual) AFact() {}

func (w warnContextual) String() string { return string(w) }

func run(pass *analysis.Pass, c *Config) (interface{}, error) {
	for _, file := range pass.Files {
		// ancestors tracks the current path from the file down to the node
		// that is currently visited. It is needed by checkForMissingLogger to
		// find the innermost enclosing function body of a call expression.
		// ast.Inspect calls the visitor function once more with a nil node
		// after it is done with the children of a node, which is used here to
		// pop that node off again.
		var ancestors []ast.Node
		ast.Inspect(file, func(n ast.Node) bool {
			if n == nil {
				ancestors = ancestors[:len(ancestors)-1]
				return true
			}
			ancestors = append(ancestors, n)

			switch n := n.(type) {
			case *ast.CallExpr:
				// We are interested in function calls, as we want to detect klog.* calls
				// passing all function calls to checkForFunctionExpr
				checkForFunctionExpr(n, pass, c)
				checkForMissingLogger(n, ancestors, pass, c)
			case *ast.FuncType:
				checkForContextAndLogger(n, n.Params, pass, c)
			case *ast.IfStmt:
				checkForIfEnabled(n, pass, c)
			case *ast.FuncDecl:
				checkForComments(pass.TypesInfo.ObjectOf(n.Name), n.Doc, pass)
			case *ast.InterfaceType:
				for _, method := range n.Methods.List {
					for _, name := range method.Names {
						checkForComments(pass.TypesInfo.ObjectOf(name), method.Doc, pass)
					}
				}
			}

			return true
		})
	}
	return nil, nil
}

// checkForFunctionExpr checks for unstructured logging function, prints error if found any.
func checkForFunctionExpr(fexpr *ast.CallExpr, pass *analysis.Pass, c *Config) {
	fun := fexpr.Fun
	args := fexpr.Args
	filename := pass.Pkg.Path() + "/" + path.Base(pass.Fset.Position(fexpr.Pos()).Filename)
	contextualCheckEnabled := c.isEnabled(contextualCheck, filename)

	// Some function that is banned for contextual logging through comment?
	if contextualCheckEnabled {
		if ident, ok := fun.(*ast.Ident); ok {
			object := pass.TypesInfo.ObjectOf(ident)
			var why warnContextual
			if pass.ImportObjectFact(object, &why) {
				pass.Report(analysis.Diagnostic{
					Pos:     fun.Pos(),
					Message: string(why),
				})
			}
		}
	}

	/* we are extracting external package function calls e.g. klog.Infof fmt.Printf
	   and eliminating calls like setLocalHost()
	   basically function calls that has selector expression like .
	*/
	if selExpr, ok := fun.(*ast.SelectorExpr); ok {
		// extracting function Name like Infof
		fName := selExpr.Sel.Name

		valueCheckEnabled := c.isEnabled(valueCheck, filename)
		keyCheckEnabled := c.isEnabled(keyCheck, filename)
		parametersCheckEnabled := c.isEnabled(parametersCheck, filename)

		// Some method that is banned for contextual logging through comment?
		if contextualCheckEnabled {
			object := pass.TypesInfo.ObjectOf(selExpr.Sel)
			var why warnContextual
			if pass.ImportObjectFact(object, &why) {
				pass.Report(analysis.Diagnostic{
					Pos:     selExpr.Sel.Pos(),
					Message: string(why),
				})
			}
		}

		// Now we need to determine whether it is coming from klog.
		if isKlog(selExpr.X, pass) {
			if c.isEnabled(contextualCheck, filename) && !isContextualCall(fName) {
				pass.Report(analysis.Diagnostic{
					Pos:     fun.Pos(),
					Message: fmt.Sprintf("function %q should not be used, convert to contextual logging", fName),
				})
				return
			}

			// Check for Deprecated function usage
			if c.isEnabled(deprecationsCheck, filename) {
				message, deprecatedUse := isDeprecatedContextualCall(fName)
				if deprecatedUse {
					pass.Report(analysis.Diagnostic{
						Pos:     fun.Pos(),
						Message: message,
					})
				}
			}

			// Matching if any unstructured logging function is used.
			if c.isEnabled(structuredCheck, filename) && isUnstructured(fName) {
				pass.Report(analysis.Diagnostic{
					Pos:     fun.Pos(),
					Message: fmt.Sprintf("unstructured logging function %q should not be used", fName),
				})
				return
			}

			// variadic input is a valid input to klog.Error*, klog.Info*, logr.Logger.Info and logr.Logger.Error
			// functions. Hence checking the parameters for variadic input argument is excluded.
			if !fexpr.Ellipsis.IsValid() {
				if keyCheckEnabled || parametersCheckEnabled || valueCheckEnabled {
					// if format specifier is used, check for arg length will most probably fail
					// so check for format specifier first and skip if found
					if parametersCheckEnabled && checkForFormatSpecifier(fexpr, pass) {
						return
					}
					switch fName {
					case "InfoS", "LoggerWithValues":
						kvCheck(args[1:], fun, pass, fName, keyCheckEnabled, parametersCheckEnabled, valueCheckEnabled)
					case "ErrorS":
						kvCheck(args[2:], fun, pass, fName, keyCheckEnabled, parametersCheckEnabled, valueCheckEnabled)
					}
				}
			}
			// verbosity Zero Check
			if c.isEnabled(verbosityZeroCheck, filename) {
				checkForVerbosityZero(fexpr, pass)
			}
		} else if isGoLogger(selExpr.X, pass) {
			if !fexpr.Ellipsis.IsValid() {
				if keyCheckEnabled || parametersCheckEnabled || valueCheckEnabled {
					// if format specifier is used, check for arg length will most probably fail
					// so check for format specifier first and skip if found
					if parametersCheckEnabled && checkForFormatSpecifier(fexpr, pass) {
						return
					}
					switch fName {
					case "WithValues":
						kvCheck(args, fun, pass, fName, keyCheckEnabled, parametersCheckEnabled, valueCheckEnabled)
					case "Info":
						kvCheck(args[1:], fun, pass, fName, keyCheckEnabled, parametersCheckEnabled, valueCheckEnabled)
					case "Error":
						kvCheck(args[2:], fun, pass, fName, keyCheckEnabled, parametersCheckEnabled, valueCheckEnabled)
					}
				}
			}
			if c.isEnabled(withHelpersCheck, filename) {
				switch fName {
				case "WithValues", "WithName":
					pass.Report(analysis.Diagnostic{
						Pos:     fun.Pos(),
						Message: fmt.Sprintf("function %q should be called through klogr.Logger%s", fName, fName),
					})
				}
			}
			// verbosity Zero Check
			if c.isEnabled(verbosityZeroCheck, filename) {
				checkForVerbosityZero(fexpr, pass)
			}

			if fName == "Error" && c.isEnabled(verbosityErrorCheck, filename) {
				// Is X itself the result of a selector expression with V as method name?
				if innerCallExpr, ok := selExpr.X.(*ast.CallExpr); ok {
					if innerSelExpr, ok := innerCallExpr.Fun.(*ast.SelectorExpr); ok && innerSelExpr.Sel.Name == "V" {
						pass.Report(analysis.Diagnostic{
							Pos:     innerSelExpr.Sel.Pos(),
							Message: `V().Error ignores the verbosity and always logs. Use only Error if that is desired, otherwise V().Info(..., "err", err).`,
						})
					}
				}
			}
		} else if fName == "NewContext" &&
			isPackage(selExpr.X, "github.com/go-logr/logr", pass) &&
			c.isEnabled(withHelpersCheck, filename) {
			pass.Report(analysis.Diagnostic{
				Pos:     fun.Pos(),
				Message: fmt.Sprintf("function %q should be called through klogr.NewContext", fName),
			})
		}

	}
}

// isKlogVerbose returns true if the type of the expression is klog.Verbose (=
// the result of klog.V).
func isKlogVerbose(expr ast.Expr, pass *analysis.Pass) bool {
	if typeAndValue, ok := pass.TypesInfo.Types[expr]; ok {
		switch t := unwrapAlias(typeAndValue.Type).(type) {
		case *types.Named:
			if typeName := t.Obj(); typeName != nil {
				if pkg := typeName.Pkg(); pkg != nil {
					if typeName.Name() == "Verbose" && pkg.Path() == "k8s.io/klog/v2" {
						return true
					}
				}
			}
		}
	}
	return false
}

// isKlog checks whether an expression is klog.Verbose or the klog package itself.
func isKlog(expr ast.Expr, pass *analysis.Pass) bool {
	// For klog.V(1) and klogV := klog.V(1) we can decide based on the type.
	if isKlogVerbose(expr, pass) {
		return true
	}

	// In "klog.Info", "klog" is a package identifier. It doesn't need to
	// be "klog" because here we look up the actual package.
	return isPackage(expr, "k8s.io/klog/v2", pass)
}

// isPackage checks whether an expression is an identifier that refers
// to a specific package like k8s.io/klog/v2.
func isPackage(expr ast.Expr, packagePath string, pass *analysis.Pass) bool {
	if ident, ok := expr.(*ast.Ident); ok {
		if object, ok := pass.TypesInfo.Uses[ident]; ok {
			switch object := object.(type) {
			case *types.PkgName:
				pkg := object.Imported()
				if pkg.Path() == packagePath {
					return true
				}
			}
		}
	}

	return false
}

// isGoLogger checks whether an expression is logr.Logger.
func isGoLogger(expr ast.Expr, pass *analysis.Pass) bool {
	if typeAndValue, ok := pass.TypesInfo.Types[expr]; ok {
		switch t := unwrapAlias(typeAndValue.Type).(type) {
		case *types.Named:
			if typeName := t.Obj(); typeName != nil {
				if pkg := typeName.Pkg(); pkg != nil {
					if typeName.Name() == "Logger" && pkg.Path() == "github.com/go-logr/logr" {
						return true
					}
				}
			}
		}
	}
	return false
}

func isUnstructured(fName string) bool {
	// List of klog functions we do not want to use after migration to structured logging.
	unstrucured := []string{
		"Infof", "Info", "Infoln", "InfoDepth",
		"Warning", "Warningf", "Warningln", "WarningDepth",
		"Error", "Errorf", "Errorln", "ErrorDepth",
		"Fatal", "Fatalf", "Fatalln", "FatalDepth",
		"Exit", "Exitf", "Exitln", "ExitDepth",
	}

	for _, name := range unstrucured {
		if fName == name {
			return true
		}
	}

	return false
}

func isDeprecatedContextualCall(fName string) (message string, deprecatedUse bool) {
	deprecatedContextualLogHelper := map[string]string{
		"KObjs": "KObjSlice",
	}
	var replacementFunction string
	if replacementFunction, deprecatedUse = deprecatedContextualLogHelper[fName]; deprecatedUse {
		message = fmt.Sprintf(`Detected usage of deprecated helper "%s". Please switch to "%s" instead.`, fName, replacementFunction)
		return
	}
	return
}

func isContextualCall(fName string) bool {
	// List of klog functions we still want to use after migration to
	// contextual logging. This is an allow list, so any new acceptable
	// klog call has to be added here.
	contextual := []string{
		"Background",
		"ClearLogger",
		"ContextualLogger",
		"EnableContextualLogging",
		"FlushAndExit",
		"FlushLogger",
		"Format",
		"FromContext",
		"InitFlags",
		"KObj",
		"KObjs",
		"KObjSlice",
		"KRef",
		"LoggerWithName",
		"LoggerWithValues",
		"NewContext",
		"SafePtr",
		"SetLogger",
		"SetLoggerWithOptions",
		"StartFlushDaemon",
		"StopFlushDaemon",
		"TODO",
	}
	for _, name := range contextual {
		if fName == name {
			return true
		}
	}

	return false
}

func checkForFormatSpecifier(expr *ast.CallExpr, pass *analysis.Pass) bool {
	if selExpr, ok := expr.Fun.(*ast.SelectorExpr); ok {
		// extracting function Name like Infof
		fName := selExpr.Sel.Name
		if strings.HasSuffix(fName, "f") {
			// Allowed for calls like Infof.
			return false
		}
		if specifier, found := hasFormatSpecifier(expr.Args); found {
			msg := fmt.Sprintf("logging function %q should not use format specifier %q", fName, specifier)
			pass.Report(analysis.Diagnostic{
				Pos:     expr.Fun.Pos(),
				Message: msg,
			})
			return true
		}
	}
	return false
}

func hasFormatSpecifier(fArgs []ast.Expr) (string, bool) {
	formatSpecifiers := []string{
		"%v", "%+v", "%#v", "%T",
		"%t", "%b", "%c", "%d", "%o", "%O", "%q", "%x", "%X", "%U",
		"%e", "%E", "%f", "%F", "%g", "%G", "%s", "%q", "%p",
	}
	for _, fArg := range fArgs {
		if arg, ok := fArg.(*ast.BasicLit); ok {
			for _, specifier := range formatSpecifiers {
				if strings.Contains(arg.Value, specifier) {
					return specifier, true
				}
			}
		}
	}
	return "", false
}

// checkForContextAndLogger ensures that a function doesn't accept both a
// context and a logger. That is problematic because it leads to ambiguity:
// does the context already contain the logger? That matters when passing it on
// without the logger.
func checkForContextAndLogger(n ast.Node, params *ast.FieldList, pass *analysis.Pass, c *Config) {
	var haveLogger, haveContext bool

	for _, param := range params.List {
		if typeAndValue, ok := pass.TypesInfo.Types[param.Type]; ok {
			switch t := unwrapAlias(typeAndValue.Type).(type) {
			case *types.Named:
				if typeName := t.Obj(); typeName != nil {
					if pkg := typeName.Pkg(); pkg != nil {
						if typeName.Name() == "Logger" && pkg.Path() == "github.com/go-logr/logr" {
							haveLogger = true
						} else if typeName.Name() == "Context" && pkg.Path() == "context" {
							haveContext = true
						}
					}
				}
			}
		}
	}

	if haveLogger && haveContext {
		pass.Report(analysis.Diagnostic{
			Pos:     n.Pos(),
			End:     n.End(),
			Message: `A function should accept either a context or a logger, but not both. Having both makes calling the function harder because it must be defined whether the context must contain the logger and callers have to follow that.`,
		})
	}
}

// checkForMissingLogger implements the optional "logger-constructor" check. It looks for
// calls to functions or methods which accept a configuration struct
// containing an optional "Logger *logr.Logger" field (or the klog.Logger
// alias for it) and checks whether the call site populates that field.
//
// This relies on a best-effort, intraprocedural data flow analysis: the
// configuration struct value passed into the call must either be a composite
// literal or a local variable which was initialized with a composite literal
// (optionally followed by "variable.Logger = ..." assignments) earlier in the
// same function. Anything else - in particular a configuration struct that
// was received as a parameter of the enclosing function or returned by some
// other call - cannot be analyzed locally, so such call sites are silently
// skipped instead of risking a false positive.
func checkForMissingLogger(call *ast.CallExpr, ancestors []ast.Node, pass *analysis.Pass, c *Config) {
	filename := pass.Pkg.Path() + "/" + path.Base(pass.Fset.Position(call.Pos()).Filename)
	if !c.isEnabled(loggerCheck, filename) || !c.isEnabled(contextualCheck, filename) {
		return
	}

	sig, ok := pass.TypesInfo.TypeOf(call.Fun).(*types.Signature)
	if !ok {
		return
	}

	params := sig.Params()
	for i := 0; i < params.Len() && i < len(call.Args); i++ {
		// The struct with the optional logger is passed by value, so the
		// variadic parameter (a slice) is never going to match.
		if sig.Variadic() && i == params.Len()-1 {
			continue
		}
		fieldIndex, ok := loggerFieldIndex(params.At(i).Type())
		if !ok {
			continue
		}

		body := enclosingFuncBody(ancestors)
		lit, fieldAssigns, ok := resolveConfigLiteral(call.Args[i], body, call.Pos(), pass)
		if !ok {
			// Cannot be checked locally.
			continue
		}

		loggerIsSet := compositeLitSetsLoggerField(lit, fieldIndex)
		if set, ok := lastLoggerFieldAssignment(fieldAssigns); ok {
			// A later "config.Logger = ..." assignment overrides whatever
			// the composite literal did.
			loggerIsSet = set
		}
		if loggerIsSet {
			continue
		}

		pass.Report(analysis.Diagnostic{
			Pos:     call.Args[i].Pos(),
			End:     call.Args[i].End(),
			Message: fmt.Sprintf("the Logger field of %s is not set, contextual logging will not work as expected unless a fallback logger is acceptable here", params.At(i).Type().String()),
		})
	}
}

// loggerFieldIndex returns the index of a directly declared "Logger" field
// with type "*logr.Logger" (or an alias of it, like klog.Logger) in t, which
// may be a struct or a pointer to one, possibly a generic instantiation.
func loggerFieldIndex(t types.Type) (int, bool) {
	t = unwrapAlias(t)
	if ptr, ok := t.(*types.Pointer); ok {
		t = unwrapAlias(ptr.Elem())
	}
	named, ok := t.(*types.Named)
	if !ok {
		return 0, false
	}
	strct, ok := named.Underlying().(*types.Struct)
	if !ok {
		return 0, false
	}
	for i := 0; i < strct.NumFields(); i++ {
		field := strct.Field(i)
		if field.Name() == "Logger" && isLoggerPointerType(field.Type()) {
			return i, true
		}
	}
	return 0, false
}

// isLoggerPointerType returns true for "*logr.Logger" and aliases of it, like
// "*klog.Logger".
func isLoggerPointerType(t types.Type) bool {
	ptr, ok := unwrapAlias(t).(*types.Pointer)
	if !ok {
		return false
	}
	named, ok := unwrapAlias(ptr.Elem()).(*types.Named)
	if !ok {
		return false
	}
	typeName := named.Obj()
	if typeName == nil {
		return false
	}
	pkg := typeName.Pkg()
	return pkg != nil && typeName.Name() == "Logger" && pkg.Path() == "github.com/go-logr/logr"
}

// enclosingFuncBody returns the body of the innermost function declaration or
// function literal in ancestors, or nil if there is none.
func enclosingFuncBody(ancestors []ast.Node) *ast.BlockStmt {
	for i := len(ancestors) - 1; i >= 0; i-- {
		switch n := ancestors[i].(type) {
		case *ast.FuncDecl:
			return n.Body
		case *ast.FuncLit:
			return n.Body
		}
	}
	return nil
}

// resolveConfigLiteral tries to determine the composite literal that was used
// to construct the value of expr and any assignments to its fields that
// happen afterwards, but still before pos, within body. It returns ok ==
// false if expr isn't a composite literal itself and isn't a simple local
// variable that can be traced back to one within the same function body,
// meaning that the value cannot be analyzed with this best-effort approach.
func resolveConfigLiteral(expr ast.Expr, body *ast.BlockStmt, pos token.Pos, pass *analysis.Pass) (*ast.CompositeLit, []*ast.AssignStmt, bool) {
	expr = unwrapAddrAndParens(expr)
	if lit, ok := expr.(*ast.CompositeLit); ok {
		return lit, nil, true
	}

	ident, ok := expr.(*ast.Ident)
	if !ok || body == nil {
		return nil, nil, false
	}
	obj := pass.TypesInfo.ObjectOf(ident)
	if obj == nil {
		return nil, nil, false
	}

	var lit *ast.CompositeLit
	var fieldAssigns []*ast.AssignStmt
	ast.Inspect(body, func(n ast.Node) bool {
		if n == nil || n.Pos() >= pos {
			return false
		}
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		if len(assign.Lhs) != len(assign.Rhs) {
			// Not a simple 1:1 assignment (for example "a, b := f()"),
			// too complex to analyze reliably.
			return true
		}
		for i, lhs := range assign.Lhs {
			switch lhs := lhs.(type) {
			case *ast.Ident:
				if pass.TypesInfo.ObjectOf(lhs) != obj {
					continue
				}
				// (Re)assignment of the variable itself. Only a composite
				// literal can be analyzed further, anything else (for
				// example a function call) makes the value untraceable
				// from here on.
				if rl, ok := unwrapAddrAndParens(assign.Rhs[i]).(*ast.CompositeLit); ok {
					lit = rl
				} else {
					lit = nil
				}
				fieldAssigns = nil
			case *ast.SelectorExpr:
				if selIdent, ok := lhs.X.(*ast.Ident); ok && pass.TypesInfo.ObjectOf(selIdent) == obj {
					fieldAssigns = append(fieldAssigns, assign)
				}
			}
		}
		return true
	})

	if lit == nil {
		return nil, nil, false
	}
	return lit, fieldAssigns, true
}

// unwrapAddrAndParens removes "&" and parentheses around expr.
func unwrapAddrAndParens(expr ast.Expr) ast.Expr {
	for {
		switch e := expr.(type) {
		case *ast.ParenExpr:
			expr = e.X
		case *ast.UnaryExpr:
			if e.Op != token.AND {
				return expr
			}
			expr = e.X
		default:
			return expr
		}
	}
}

// compositeLitSetsLoggerField returns true if the composite literal sets the
// field at fieldIndex (the "Logger" field) to something other than nil,
// either through a keyed or a positional element.
func compositeLitSetsLoggerField(lit *ast.CompositeLit, fieldIndex int) bool {
	keyed := false
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		keyed = true
		if ident, ok := kv.Key.(*ast.Ident); ok && ident.Name == "Logger" {
			return !isNilIdent(kv.Value)
		}
	}
	if keyed {
		// Fully keyed literal without a "Logger" key: not set.
		return false
	}
	// Positional literal: Go requires that either all or none of the
	// elements are keyed, so at this point all elements are positional.
	if fieldIndex < len(lit.Elts) {
		return !isNilIdent(lit.Elts[fieldIndex])
	}
	return false
}

// lastLoggerFieldAssignment returns whether the last assignment in
// fieldAssigns which targets the "Logger" field sets it to something other
// than nil. ok is false if none of the assignments target that field.
func lastLoggerFieldAssignment(fieldAssigns []*ast.AssignStmt) (set bool, ok bool) {
	for _, assign := range fieldAssigns {
		for i, lhs := range assign.Lhs {
			selExpr, selOK := lhs.(*ast.SelectorExpr)
			if !selOK || selExpr.Sel.Name != "Logger" {
				continue
			}
			set = !isNilIdent(assign.Rhs[i])
			ok = true
		}
	}
	return set, ok
}

// isNilIdent returns true if expr is the predeclared identifier "nil".
func isNilIdent(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == "nil"
}

// checkForIfEnabled detects `if klog.V(..).Enabled() { ...` and `if
// logger.V(...).Enabled()` and suggests capturing the result of V.
func checkForIfEnabled(i *ast.IfStmt, pass *analysis.Pass, c *Config) {
	// if i.Init == nil {
	// A more complex if statement, let's assume it's okay.
	// return
	// }

	// Must be a method call.
	callExpr, ok := i.Cond.(*ast.CallExpr)
	if !ok {
		return
	}
	selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}

	// We only care about calls to Enabled().
	if selExpr.Sel.Name != "Enabled" {
		return
	}

	// And it must be Enabled for klog or logr.Logger.
	if !isKlogVerbose(selExpr.X, pass) &&
		!isGoLogger(selExpr.X, pass) {
		return
	}

	// logger.Enabled() is okay, logger.V(1).Enabled() is not.
	// That means we need to check for another selector expression
	// with V as method name.
	subCallExpr, ok := selExpr.X.(*ast.CallExpr)
	if !ok {
		return
	}
	subSelExpr, ok := subCallExpr.Fun.(*ast.SelectorExpr)
	if !ok || subSelExpr.Sel.Name != "V" {
		return
	}

	// klogV is recommended as replacement for klog.V(). For logr.Logger
	// let's use the root of the selector, which should be a variable.
	varName := "klogV"
	funcCall := "klog.V"
	if isGoLogger(subSelExpr.X, pass) {
		varName = "logger"
		root := subSelExpr
		for s, ok := root.X.(*ast.SelectorExpr); ok; s, ok = root.X.(*ast.SelectorExpr) {
			root = s
		}
		if id, ok := root.X.(*ast.Ident); ok {
			varName = id.Name
		}
		funcCall = varName + ".V"
	}

	pass.Report(analysis.Diagnostic{
		Pos: i.Pos(),
		End: i.End(),
		Message: fmt.Sprintf("the result of %s should be stored in a variable and then be used multiple times: if %s := %s(); %s.Enabled() { ... %s.Info ... }",
			funcCall, varName, funcCall, varName, varName),
	})
}

func checkForVerbosityZero(fexpr *ast.CallExpr, pass *analysis.Pass) {
	iselExpr, ok := fexpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}
	expr := iselExpr.X
	if !isKlogVerbose(expr, pass) && !isGoLogger(expr, pass) {
		return
	}
	if isVerbosityZero(expr) {
		msg := "Logging with V(0) is semantically equivalent to the same expression without it and just causes unnecessary overhead. It should get removed."
		pass.Report(analysis.Diagnostic{
			Pos:     fexpr.Fun.Pos(),
			Message: msg,
		})
	}
}

func isVerbosityZero(expr ast.Expr) bool {
	subCallExpr, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	subSelExpr, ok := subCallExpr.Fun.(*ast.SelectorExpr)
	if !ok || subSelExpr.Sel.Name != "V" || len(subCallExpr.Args) != 1 {
		return false
	}

	if lit, ok := subCallExpr.Args[0].(*ast.BasicLit); ok {
		return lit.Value == "0"
	}

	// When Constants of value is defined in different files, the id.Obj will be nil, we should filter this condition.
	if id, ok := subCallExpr.Args[0].(*ast.Ident); ok && id.Obj != nil && id.Obj.Kind == 2 {
		v, ok := id.Obj.Decl.(*ast.ValueSpec)
		if !ok || len(v.Values) != 1 {
			return false
		}
		if lit, ok := v.Values[0].(*ast.BasicLit); ok && lit.Value == "0" {
			return true
		}
	}
	return false
}

// kvCheck check if all keys in keyAndValues are valid keys according to the guidelines
// and that the values can be formatted.
func kvCheck(keyValues []ast.Expr, fun ast.Expr, pass *analysis.Pass, funName string, keyCheckEnabled, parametersCheckEnabled, valueCheckEnabled bool) {
	if len(keyValues)%2 != 0 {
		pass.Report(analysis.Diagnostic{
			Pos:     fun.Pos(),
			Message: fmt.Sprintf("Additional arguments to %s should always be Key Value pairs. Please check if there is any key or value missing.", funName),
		})
		return
	}

	for index, arg := range keyValues {
		switch index % 2 {
		case 0:
			// Key in key/value pair.
			checkKey(arg, pass, keyCheckEnabled, parametersCheckEnabled)
		case 1:
			// Value in key/value pair.
			checkValue(arg, pass, valueCheckEnabled)
		}
	}
}

// checkKey checks the key in a key/value pair.
func checkKey(arg ast.Expr, pass *analysis.Pass, keyCheckEnabled, parametersCheckEnabled bool) {
	if !keyCheckEnabled && !parametersCheckEnabled {
		return
	}

	lit, ok := arg.(*ast.BasicLit)
	if !ok {
		pass.Report(analysis.Diagnostic{
			Pos:     arg.Pos(),
			Message: fmt.Sprintf("Key positional arguments are expected to be inlined constant strings. Please replace %v provided with string value.", arg),
		})
		return
	}

	if lit.Kind != token.STRING {
		pass.Report(analysis.Diagnostic{
			Pos:     arg.Pos(),
			Message: fmt.Sprintf("Key positional arguments are expected to be inlined constant strings. Please replace %v provided with string value.", lit.Value),
		})
		return
	}

	switch {
	case parametersCheckEnabled:
		// This is the less strict check.
		isASCII := utf8string.NewString(lit.Value).IsASCII()
		if !isASCII {
			pass.Report(analysis.Diagnostic{
				Pos:     arg.Pos(),
				Message: fmt.Sprintf("Key positional arguments %s are expected to be lowerCamelCase alphanumeric strings. Please remove any non-Latin characters.", lit.Value),
			})
		}
	case keyCheckEnabled:
		// This is the stricter check.
		keyMatchRe := regexp.MustCompile(`(^[A-Z]{2,}|^[a-z])[[:alnum:]]*$`)
		match := keyMatchRe.Match([]byte(strings.Trim(lit.Value, "\"")))
		if !match {
			pass.Report(analysis.Diagnostic{
				Pos:     arg.Pos(),
				Message: fmt.Sprintf("Key positional arguments %s are expected to be alphanumeric and start with either one lowercase or two uppercase letters. Please refer to https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments.", lit.Value),
			})
		}
	}
}

// checkValue checks the value in a key/value pair.
func checkValue(arg ast.Expr, pass *analysis.Pass, valueCheckEnabled bool) {
	if !valueCheckEnabled {
		return
	}

	// Check the type.
	if typeAndValue, ok := pass.TypesInfo.Types[arg]; ok {
		if obj, index, _ := types.LookupFieldOrMethod(typeAndValue.Type, typeAndValue.Addressable(), nil /* package */, "String"); obj != nil {
			if function, ok := obj.(*types.Func); ok && isFmtString(function) && len(index) > 1 && !isWrapperStruct(typeAndValue.Type) {
				pass.Report(analysis.Diagnostic{
					Pos:     arg.Pos(),
					Message: fmt.Sprintf("The type %s inherits %s as implementation of fmt.Stringer, which covers only a subset of the value. Implement String() for the type or wrap it with TODO.", typeAndValue.Type.String(), function.FullName()), // TODO: https://github.com/kubernetes/kubernetes/pull/116952
				})
			}
		}
	}
}

// isFmtString checks whether the function has the "func() string" signature.
func isFmtString(function *types.Func) bool {
	signature, ok := function.Type().(*types.Signature)
	if !ok {
		return false
	}
	params := signature.Params()
	if params != nil && params.Len() != 0 {
		return false
	}
	results := signature.Results()
	if results == nil || results.Len() != 1 {
		return false
	}
	result := results.At(0)
	basic, ok := result.Type().(*types.Basic)
	if !ok {
		return false
	}
	if basic.Kind() != types.String {
		return false
	}
	return true
}

// isWrapperStruct returns true for types that are a struct with a single field
// or a pointer to one, like for example
// https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Time.
func isWrapperStruct(t types.Type) bool {
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	if named, ok := t.(*types.Named); ok {
		t = named.Underlying()
	}
	if strct, ok := t.(*types.Struct); ok {
		return strct.NumFields() == 1
	}

	return false
}

func checkForComments(object types.Object, doc *ast.CommentGroup, pass *analysis.Pass) {
	if object == nil || doc == nil {
		return
	}

	for _, comment := range doc.List {
		text := comment.Text
		text, found := strings.CutPrefix(text, logcheckPrefix)
		if !found {
			continue
		}
		text, found = strings.CutPrefix(text, contextKeyword)
		if !found {
			pass.Report(analysis.Diagnostic{
				Pos:     comment.Pos(),
				Message: "unknown logcheck keyword in comment",
			})
			continue
		}
		text = strings.TrimSpace(text)
		why := warnContextual(fmt.Sprintf("%s should not be used in code which supports contextual logging.", object.Name()))
		text, found = strings.CutPrefix(text, commentSep)
		if found {
			text = strings.TrimSpace(text)
			if len(text) > 0 {
				why = warnContextual(text)
			}
		}
		pass.ExportObjectFact(object, &why)
	}
}

func unwrapAlias(t types.Type) types.Type {
	switch t := t.(type) {
	case *types.Alias:
		return unwrapAlias(t.Rhs())
	default:
		return t
	}
}

const (
	logcheckPrefix = "//logcheck:"
	contextKeyword = "context"
	commentSep     = "//"
)
