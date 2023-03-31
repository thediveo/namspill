# namspill

[![PkgGoDev](https://img.shields.io/badge/-reference-blue?logo=go&logoColor=white&labelColor=505050)](https://pkg.go.dev/github.com/thediveo/namspill)
[![GitHub](https://img.shields.io/github/license/thediveo/namspill)](https://img.shields.io/github/license/thediveo/namspill)
[![Go Report Card](https://goreportcard.com/badge/github.com/thediveo/namspill)](https://goreportcard.com/report/github.com/thediveo/namspill)
![Coverage](https://img.shields.io/badge/Coverage-100.0%25-brightgreen)

> ℹ️ If you don't switch Linux-kernel namespaces and don't use a Go module that
> does, you most probably don't need to fret about `namspill`. 

In the spirit of leak tests for
[goroutines](https://onsi.github.io/gomega/#codegleakcode-finding-leaked-goroutines)
and [file descriptors](https://github.com/thediveo/fdooze), `namspill` tests for
Linux kernel namespaces unintendedly "leaking" between goroutines due to
incorrect [OS thread locking](https://pkg.go.dev/runtime#LockOSThread) (or
rather, the lack thereof) when switching namespaces in a Go program.

`namspill` is primarily designed to integrate smoothly with the
[Ginkgo](https://github.com/onsi/ginkgo) BDD testing framework and the
[Gomega](https://github.com/onsi/gomega) matcher/assertion library. (It may be
used also outside Ginkgo/Gomega, but such usage is out of scope.)

## Install

```bash
go get github.com/thediveo/namspill
```

## Usage

In its simplest form, just after each test gather information about the
namespaces of the tasks (threads) of the program and check that no namespaces
have leaked, with some tasks being attached to other namespaces than the rest of
the tasks.

As `namspill` exports only very few symbols and is intended to extend the Gomega
DSL, you might want to dot-import it.

```go
import . "github.com/thediveo/namspill"

AfterEach(func() {
    // You might want to use Eventually(Tasks)... in case you don't
    // have a preceeding Eventually(Goroutines)... that waits for
    // the Goroutines (and thus Linux threads/tasks) to settle first.
    Expect(Tasks()).To(BeUniformlyNamespaced())
})
```

Normally, you shouldn't then see anything unless there is a problem with threads
attached to other Linux-kernel namespaces when they shouldn't. In this case, the
`BeUniformlyNamespaced` matcher will fail and show you the list of tasks with
the namespaces they're attached to.

Unfortunately, there's no way to show you which goroutine might have caused this
nor the call site (and more importantly, the call stack) where the namespace
switch happened.

## Background

`namspill` mostly serves as a canary to (quickly) detect forgetting to switch
back namespaces before unlocking OS threads so that they can be freely scheduled
to any arbitrary goroutine.

### The Initial Thread M0 _Is_ Special

Additionally, `namespill` checks safeguard against alternatively not terminating
thread-locked goroutines. And that is where there's a catch when it comes to the
Go scheduler: if a Go program's initial thread ("M0" in Go scheduler parlance)
ends up being locked to a (non-main) goroutine and this goroutine exits ... then
the initial thread doesn't get terminated, because that usually has unwanted
side-effects on several operating systems. Instead, the Go scheduler "wedges"
the initial thread and never schedules any goroutine to it again.

This situation will (correctly) trigger failed `namspill` assertions. To avoid
the initial thread getting wedged, simply lock it in an `init` function to the
initial/main goroutine, so it never ends up getting scheduled on any other
goroutine (that might be subjected to locking it to a thread and then
terminating it while being locked):

```go
func init() {
    runtime.LockOSThread()
}
```

## References

For further background information, please see the following references:

- [M0 _is_ Special](http://thediveo.github.io/#/art/namspill) – many more
  background details and how the pieces fit together.

- [LockOSThread, switching (Linux kernel) namespaces: what happens to the main
  thread...?](https://groups.google.com/g/golang-nuts/c/dx-jweSVxHk) – my
  original question and the following highly useful discussion in the
  golang-nuts group drilling down into the scheduler behavior when it comes to
  the asymmetry between the initial thread (thread group leader) and its fellow
  threads.

- [runtime: on Linux, better do not treat the initial thread/task group leader
  as any other thread/task](https://github.com/golang/go/issues/53210) – points
  the finger to the scheduler source code where the initial thread, a.k.a. "M0",
  turns out to be special after all.

## Make Targets

- `make`: lists all targets.
- `make coverage`: runs all tests with coverage and then **updates the coverage
  badge in `README.md`**.
- `make pkgsite`: installs [`x/pkgsite`](golang.org/x/pkgsite/cmd/pkgsite), as
  well as the [`browser-sync`](https://www.npmjs.com/package/browser-sync) and
  [`nodemon`](https://www.npmjs.com/package/nodemon) npm packages first, if not
  already done so. Then runs the `pkgsite` and hot reloads it whenever the
  documentation changes.
- `make report`: installs
  [`@gojp/goreportcard`](https://github.com/gojp/goreportcard) if not yet done
  so and then runs it on the code base.
- `make test`: runs all tests.

## Copyright and License

`namspill` is Copyright 2022, 2022 Harald Albrecht, and licensed under the
Apache License, Version 2.0.
