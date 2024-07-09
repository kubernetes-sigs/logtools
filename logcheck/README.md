This directory contains a linter for checking log calls. It was originally
created to detect when unstructured logging calls like `klog.Infof` get added
to files that should only use structured logging calls like `klog.InfoS`
and now also supports other checks.

# Installation

`go install sigs.k8s.io/logtools/logcheck`

# Usage

`$logcheck.go <package_name>`
`e.g $logcheck ./pkg/kubelet/lifecycle/`

# Configuration

Checks can be enabled or disabled globally via command line flags and env
variables. In addition, the global setting for a check can be modified per file
via a configuration file. That file contains lines in this format:

```
<checks> <regular expression>
```

`<checks>` is a comma-separated list of the names of checks that get enabled or
disabled when a file name matches the regular expression. A check gets disabled
when its name has `-` as prefix and enabled when there is no prefix or `+` as
prefix. Only checks that are mentioned explicitly are modified. All regular
expressions are checked in order, so later lines can override the previous
ones.

In this example, checking for klog calls is enabled for all files under
`pkg/scheduler` in the Kubernetes repo except for `scheduler.go`
itself. Parameter checking is disabled everywhere.

```
klog,-parameters k8s.io/kubernetes/pkg/scheduler/.*
-klog k8s.io/kubernetes/pkg/scheduler/scheduler.go
```

The names of all supported checks are the ones used as sub-section titles in
the next section.

# Checks

## structured (enabled by default)

Unstructured klog logging calls are flagged as error.

## contextual (disabled by default)

None of the klog logging methods may be used. This is even stricter than
`unstructured`. Instead, code should retrieve a logr.Logger from klog and log
through that.

klog calls that are needed to manage contextual logging, for example
`klog.Background`, are still allowed.

Which of the klog functions are allowed is compiled into the logcheck binary.
For functions or methods defined elsewhere, a special `//logcheck:context` can
be added to trigger a warning about usage of such an API when contextual
checking is enabled. Here is an example:

    //logcheck:context // Foo should not be used in code which supports contextual logging.
    func Foo() { ... }

The additional explanation is optional. The default is the text above. It is recommended
to mention what should be used instead, for example like this:

    //logcheck:context // FooWithContext should be used instead of Foo in code which supports contextual logging.
    func Foo() { ... }
    
    func FooWithContext(ctx context.Context) { .... }

## parameters (enabled by default)

This ensures that if certain logging functions are allowed and are used, those
functions are passed correct parameters.

### all calls

Format strings are not allowed where plain strings are expected.

### structured logging calls

Key/value parameters for logging calls are checked:

- For each key there must be a value.
- Keys must be constant strings.

This also warns about code that is valid, for example code that collects
key/value pairs in an `[]interface` variable before passing that on to a log
call. Such valid code can use `nolint:logcheck` to disable the warning (when
invoking logcheck through golangci-lint) or the `parameters` check can be
disabled for the file.

## with-helpers (disabled by default)

`logr.Logger.WithName`, `logr.Logger.WithValues` and `logr.NewContext` must not
be used.  The corresponding helper calls from `k8s.io/klogr` should be used
instead. This is relevant when support contextual logging is disabled at
runtime in klog.

## verbosity-zero (enabled by default)

This check flags all invocation of `klog.V(0)` or any of it's equivalent as errors

## key (enabled by default)

This check flags check whether name arguments are valid keys according to the
[Kubernetes guidelines](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#name-arguments).

## deprecations (enabled by default)

This checks detects the usage of deprecated `klog` helper functions such as `KObjs` and suggests
a suitable alternative to replace them with.

# Golangci-lint

Logcheck needs to be built as a plugin to golangci-lint to be executed as a
private linter. There are two plugin systems in golangci-lint, and the following
instructions apply to the [Module Plugin System](https://golangci-lint.run/plugins/module-plugins/)
(introduced since v1.57.0), which is a supported approach to run Logcheck in golangci-lint.

One will have to build a custom golangci-lint binary by doing:

(1) Create a `.custom-gcl.yml` file at the root of the repository if you have not
done so, add the following content:

```yaml
version: v1.59.1
plugins:
    - module: "sigs.k8s.io/logtools"
      import: "sigs.k8s.io/logtools/logcheck/gclplugin"
      version: latest
```

(2) Add logcheck to the linter configuration file `.golangci.yaml`:

```yaml
# This config disables all other linters and only runs logcheck
# Configure the linters as all other linters. This is mostly
# for brevity.
linters:
  disable-all: true
  enable:
    - logcheck

linters-settings:
  custom:
    logcheck:
      type: "module"
      description: structured logging checker
      original-url: sigs.k8s.io/logtools/logcheck
      settings:
        check:
          contextual: true
        config: |
          structured .*
          contextual .*

```

(3) Build a custom golangci-lint binary with logcheck included:

```sh
golangci-lint custom
```

(4) Run the custom binary instead of golangci-lint:

```sh
# Arguments are the same as `golangci-lint`.
./custom-gcl run ./...
```
