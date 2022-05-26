// Copyright 2022 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy
// of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package namspill

import (
	"fmt"
	"os"

	"github.com/thediveo/lxkns/nstest"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/namspill/task"
	"github.com/thediveo/testbasher"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("uniform namespacing", func() {

	It("is normal uniformly namespaced", func() {
		m := BeUniformlyNamespaced()
		success, err := m.Match(Tasks())
		Expect(err).NotTo(HaveOccurred())
		Expect(success).To(BeTrue())
	})

	It("rejects wrong types of actual values", func() {
		m := BeUniformlyNamespaced()
		Expect(m.Match(nil)).Error().To(MatchError(MatchRegexp(`expects a \[\]Task.  Got:\n\s+<nil>: nil`)))
		Expect(m.Match(42)).Error().To(MatchError(MatchRegexp(`expects a \[\]Task.  Got:\n\s+<int>: 42`)))
		Expect(m.Match([]task.Task{})).Error().To(MatchError(MatchRegexp(`expects a non-empty \[\]Task.  Got:\n\s+<\[\]task.Task \| len:0, cap:0>: \[\]`)))
	})

	It("does not succeed for different numbers of namespaces", func() {
		Expect([]task.Task{
			{TID: 1, Leader: 0, Namespaces: map[string]uint64{"foo": uint64(1)}},
			{TID: 2, Leader: 1, Namespaces: map[string]uint64{"foo": uint64(1), "bar": uint64(2)}},
		}).NotTo(BeUniformlyNamespaced())
	})

	It("does not succeed for different namespaces", func() {
		Expect([]task.Task{
			{TID: 1, Leader: 0, Namespaces: map[string]uint64{"foo": uint64(1)}},
			{TID: 2, Leader: 1, Namespaces: map[string]uint64{"bar": uint64(1)}},
		}).NotTo(BeUniformlyNamespaced())
	})

	It("does not succeed without a task leader", func() {
		Expect([]task.Task{
			{TID: 1, Leader: 2, Namespaces: map[string]uint64{"foo": uint64(1)}},
			{TID: 2, Leader: 1, Namespaces: map[string]uint64{"bar": uint64(1)}},
		}).NotTo(BeUniformlyNamespaced())
	})

	It("returns failure messages", func() {
		m := BeUniformlyNamespaced()
		Expect(m.FailureMessage(nil)).To(MatchRegexp(`Expected\n\s+\[\]task.Task\nGot:\n\s+<nil>`))
		Expect(m.NegatedFailureMessage(nil)).To(MatchRegexp(`Expected\n\s+\[\]task.Task\nGot:\n\s+<nil>`))
		ts := []task.Task{
			{TID: 1, Leader: 0, Namespaces: map[string]uint64{"foo": uint64(1)}},
			{TID: 2, Leader: 1, Namespaces: map[string]uint64{"foo": uint64(1), "bar": uint64(2)}},
		}
		Expect(m.FailureMessage(ts)).To(MatchRegexp(
			`Expected\n\s+Task Leader PID: 1, foo:\[1\]\n\s+Task TID: 2, bar:\[2\], foo:\[1\]\nto have uniform namespace IDs per task`))
		Expect(m.NegatedFailureMessage(ts)).To(MatchRegexp(
			`Expected\n\s+Task Leader PID: 1, foo:\[1\]\n\s+Task TID: 2, bar:\[2\], foo:\[1\]\nnot to have uniform namespace IDs per task`))
	})

	It("detects non-uniform namespacing", func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		By("creating a new network namespace")
		scripts := testbasher.Basher{}
		defer scripts.Done()
		scripts.Common(nstest.NamespaceUtilsScript)
		scripts.Script("main", `unshare -Unr $stage2`)
		scripts.Script("stage2", `
echo $$
read
`)
		cmd := scripts.Start("main")
		defer cmd.Close()
		var unsharedpid int
		cmd.Decode(&unsharedpid)
		Expect(unsharedpid).NotTo(BeZero())

		By("switching a separate goroutine into new network namespace")
		switched := make(chan struct{})
		done := make(chan struct{})
		unswitched := make(chan struct{})
		newnetns := ops.NewTypedNamespacePath(
			fmt.Sprintf("/proc/%d/ns/net", unsharedpid),
			species.CLONE_NEWNET,
		)
		go func() {
			defer GinkgoRecover()
			By("about to visit in new goroutine")
			err := ops.Visit(func() {
				defer GinkgoRecover()
				By("visiting goroutine")
				close(switched)
				Eventually(done).Should(BeClosed())
			}, newnetns)
			Expect(err).NotTo(HaveOccurred())
			close(unswitched)
		}()
		By("waiting for switching done")
		Eventually(switched).Should(BeClosed())

		By("checking task namespacing")
		m := BeUniformlyNamespaced()
		success, err := m.Match(Tasks())
		Expect(err).NotTo(HaveOccurred())
		Expect(success).To(BeFalse())

		By("restoring things")
		close(done)
		Eventually(unswitched).Should(BeClosed())
		Expect(Tasks()).To(BeUniformlyNamespaced())
	})

})
