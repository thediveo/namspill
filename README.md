# Namspill

In the spirit of leak tests for
[goroutines](https://onsi.github.io/gomega/#codegleakcode-finding-leaked-goroutines)
and [file descriptors](https://github.com/thediveo/fdooze), `namspill` tests for
Linux kernel namespaces unintendedly "leaking" between goroutines due to
incorrect [OS thread locking](https://pkg.go.dev/runtime#LockOSThread) (or rather, the lack thereof) when switching namespaces in a Go program.

If you don't switch Linux-kernel namespaces or use a Go module that does, you most probably don't need to fret about `namspill`. 

`namspill` is primarily designed to integrate smoothly with the
[Ginkgo](https://github.com/onsi/ginkgo) BDD testing framework and the
[Gomega](https://github.com/onsi/gomega) matcher/assertion library. It may be
used also outside Ginkgo/Gomega.

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

## Background

`namspill` mostly serves as a canary to quickly detect forgetting to switch back
namespaces before unlocking OS threads so that they can be freely scheduled to
any arbitrary goroutine.

Additionally, it safeguards against alternatively not terminating thread-locked
goroutines. And that is where there's a catch when it comes to the Go scheduler:
if a Go program's initial thread ends up being locked to a (non-main) goroutine
and this goroutine exits ... then the initial thread doesn't get terminated,
because that usually has unwanted side-effects on several operating systems.
Instead, the Go scheduler "wedges" the initial thread and never schedules any
goroutine to it again.

This situation not least will (correctly) trigger failed `namspill` assertions.
To avoid the initial thread getting wedged, simply lock it in an `init` function
to the initial/main goroutine, so it never ends up getting scheduled on any
other goroutine (that might be subjected to locking it to a thread and then
terminating it while being locked):

```go
func init() {
    runtime.LockOSThread()
}
```

For further background information, please see the following references...

- [LockOSThread, switching (Linux kernel) namespaces: what happens to the main
  thread...?](https://groups.google.com/g/golang-nuts/c/dx-jweSVxHk)

- [runtime: on Linux, better do not treat the initial thread/task group leader
  as any other thread/task](https://github.com/golang/go/issues/53210) – points
  the finger to the scheduler source code where the initial thread, a.k.a. "M0",
  turns out to be special after all.

