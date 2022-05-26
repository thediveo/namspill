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
	"sort"
	"strings"

	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
	"github.com/thediveo/namspill/task"
)

// BeUniformlyNamespaced succeeds if the actual value is a slice of Task
// elements and all tasks share the same namespaces (per namespace type).
func BeUniformlyNamespaced() types.GomegaMatcher {
	return &beUniformlyNamespacedMatcher{}
}

type beUniformlyNamespacedMatcher struct{}

func (matcher *beUniformlyNamespacedMatcher) Match(actual interface{}) (bool, error) {
	tasks, ok := actual.([]task.Task)
	if !ok {
		return false, fmt.Errorf("BeUniform matcher expects a []Task.  Got:\n%s",
			format.Object(actual, 1))
	}
	if len(tasks) == 0 {
		return false, fmt.Errorf("BeUniform matcher expects a non-empty []Task.  Got:\n%s",
			format.Object(actual, 1))
	}
	leader := leaderTask(tasks)
	for _, task := range tasks {
		if task.TID == leader.TID {
			continue
		}
		if len(task.Namespaces) != len(leader.Namespaces) {
			return false, nil
		}
		for typ, nsid := range leader.Namespaces {
			if task.Namespaces[typ] != nsid {
				return false, nil
			}
		}
	}
	return true, nil
}

func (matcher *beUniformlyNamespacedMatcher) FailureMessage(actual interface{}) (message string) {
	if tasks, ok := actual.([]task.Task); ok {
		return fmt.Sprintf("Expected\n%s\nto have uniform namespace IDs per task",
			formatTasks(tasks, 1))
	}
	return fmt.Sprintf("Expected\n%s[]task.Task\nGot:\n%s%T",
		format.Indent, format.Indent, actual)
}

func (matcher *beUniformlyNamespacedMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	if tasks, ok := actual.([]task.Task); ok {
		return fmt.Sprintf("Expected\n%s\nnot to have uniform namespace IDs per task",
			formatTasks(tasks, 1))
	}
	return fmt.Sprintf("Expected\n%s[]task.Task\nGot:\n%s%T",
		format.Indent, format.Indent, actual)
}

// formats a list of tasks in a more specific way than format.Object by sorting
// the Tasks with the leader task first and then all other tasks in TID-sorted
// order. Additionally, it uses Task's concise String representation.
func formatTasks(tasks []task.Task, indentation uint) string {
	indent := strings.Repeat(format.Indent, int(indentation))
	var s strings.Builder
	// Always render the leader task first.
	leader := leaderTask(tasks)
	s.WriteString(indent)
	s.WriteString(leader.String())
	// And now for the remaining tasks, sorted.
	tasks = append(tasks[0:0:0], tasks...)
	sort.Slice(tasks, func(a, b int) bool { return tasks[a].TID < tasks[b].TID })
	for _, task := range tasks {
		if task.Leader == 0 {
			continue
		}
		s.WriteRune('\n')
		s.WriteString(indent)
		s.WriteString(task.String())
	}
	return s.String()
}

// leaderTask returns the leader task from a list of tasks.
func leaderTask(tasks []task.Task) task.Task {
	for _, task := range tasks {
		if task.Leader == 0 {
			return task
		}
	}
	return task.Task{}
}
