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

package task

import (
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("task discovery", func() {

	It("discovers the tasks of this process", func() {
		l, err := net.Listen("tcp", "localhost:0")
		Expect(err).NotTo(HaveOccurred())
		defer l.Close()

		go func() {
			defer GinkgoRecover()
			_, err := l.Accept()
			Expect(err).To(HaveOccurred()) // sic!
		}()

		tids, leadertid := GetTaskIds()
		Expect(len(tids)).To(BeNumerically(">=", 2))
		Expect(leadertid).NotTo(BeZero())
		Expect(tids).To(HaveEach(Not(BeZero())))
		for idx, tid := range tids {
			Expect(append(tids[0:idx:idx], tids[idx+1:]...)).NotTo(ContainElement(tid),
				"TID %d contained more than once in %v", tid, tids)
		}
		Expect(tids).To(ContainElement(leadertid))

		tasks := Tasks()
		Expect(len(tasks)).To(BeNumerically(">=", 2))
		Expect(tasks).To(ContainElement(HaveField("Leader", uint32(0))))
		Expect(tasks).To(HaveEach(HaveField("TID", BeElementOf(tids))))
	})

	It("fails correctly on trying to discover TIDs", func() {
		tasks, leader := getTaskIds("./test", 1)
		Expect(tasks).To(BeNil())
		Expect(leader).To(BeZero())
	})

	It("skips invalid task dir entries", func() {
		tasks, leader := getTaskIds("./test", 10)
		Expect(tasks).To(ContainElements(uint32(1)))
		Expect(leader).To(Equal(uint32(10)))

	})

	It("fails correctly with invalid tasks", func() {
		Expect(tasks("./test")).To(BeNil())
	})

	It("prints a Task", func() {
		tids, leadertid := GetTaskIds()
		t := newTask("", leadertid, leadertid)
		Expect(t.String()).To(MatchRegexp(
			`^Task Leader PID: [1-9]\d*, ((cgroup|ipc|mnt|net|pid|time|user|uts):\[\d+\])+`))
		var noleadertids []uint32
		Expect(tids).To(ContainElement(Not(Equal(leadertid)), &noleadertids))
		t = newTask("", noleadertids[0], leadertid)
		Expect(t.String()).To(MatchRegexp(
			`^Task TID: [1-9]\d*, ((cgroup|ipc|mnt|net|pid|time|user|uts):\[\d+\])+`))
	})

	It("returns zero Task for invalid TID", func() {
		Expect(newTask("", 0, 0)).To(BeZero())
		Expect(newTask("./test", 1, 0)).To(BeZero())
		Expect(newTask("./test", 2, 0)).To(BeZero())
		Expect(newTask("./test", 3, 0)).To(BeZero())
	})

})
