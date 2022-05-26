# Namspill

In the spirit of
[goroutine](https://onsi.github.io/gomega/#codegleakcode-finding-leaked-goroutines)
and [file descriptor](https://github.com/thediveo/fdooze) leak testers,
`namspill` tests for Linux kernel namespaces leaking between uninvolved
goroutines due to incorrect [OS thread
locking](https://pkg.go.dev/runtime#LockOSThread) and namespace switching.

`namspill` is primarily designed to integrate smoothly with the
[Ginkgo](https://github.com/onsi/ginkgo) and
[Gomega](https://github.com/onsi/gomega) TDD modules. However, it can also be
used outside Ginkgo/Gomega.

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
    Expect(Tasks()).To(BeUniformlyNamespaced())
})
```
