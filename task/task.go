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
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Task represents an individual Linux task (thread) of the running Go process.
type Task struct {
	TID        uint32            // identifier of this task.
	Leader     uint32            // identifier of task leader or zero, if this is the task leader.
	Namespaces map[string]uint64 // the types and identifiers of the namespaces this task is attached to.
}

// IsTask returns true, if this Task is a zero value task and thus doesn't
// represent any alive task.
func (t Task) IsZero() bool {
	return t.TID == 0
}

// String returns a textual single-line representation of a task (leader) and
// the types and IDs of the namespaces it is attached to.
func (t Task) String() string {
	var s strings.Builder
	if t.Leader == 0 {
		s.WriteString("Task Leader PID: ")
		s.WriteString(strconv.FormatUint(uint64(t.TID), 10))
	} else {
		s.WriteString("Task TID: ")
		s.WriteString(strconv.FormatUint(uint64(t.TID), 10))
	}
	types := make([]string, 0, len(t.Namespaces))
	for typ := range t.Namespaces {
		types = append(types, typ)
	}
	sort.Strings(types)
	for _, typ := range types {
		s.WriteString(", ")
		s.WriteString(typ)
		s.WriteString(":[")
		s.WriteString(strconv.FormatUint(t.Namespaces[typ], 10))
		s.WriteRune(']')
	}
	return s.String()
}

// Tasks returns information about the tasks currently belonging to this
// process. If the tasks cannot properly be determined, Tasks returns nil
// instead.
func Tasks() []Task {
	return tasks("")
}

// tasks returns information about the tasks currently belonging to this
// process, reading task information from either a real procfs instance or a
// fake procfs surrogate for testing purposes.
func tasks(procroot string) []Task {
	tids, leader := GetTaskIds()
	tasks := make([]Task, 0, len(tids))
	for _, tid := range tids {
		task := newTask(procroot, tid, leader)
		if task.IsZero() {
			// this task has already terminated while we're gathering the
			// details, so we have to ignore it.
			continue
		}
		tasks = append(tasks, task)
	}
	return tasks
}

// newTask returns a Task for the taks with the given task ID (TID) and task
// leader. If the specified TID doesn't exist anymore, newTask returns a Task
// zero value.
func newTask(procroot string, TID uint32, leader uint32) Task {
	namespaceItems, err := os.ReadDir(fmt.Sprintf(procroot+"/proc/%d/ns", TID))
	if err != nil {
		return Task{}
	}
	namespaces := map[string]uint64{}
	for _, nsitem := range namespaceItems {
		nsidtext, err := os.Readlink(fmt.Sprintf(procroot+"/proc/%d/ns/%s", TID, nsitem.Name()))
		if err != nil {
			return Task{}
		}
		fields := strings.Split(nsidtext, ":")
		if len(fields) != 2 {
			return Task{}
		}
		nsid, err := strconv.ParseUint(strings.Trim(fields[1], "[]"), 10, 64)
		if err != nil {
			return Task{}
		}
		namespaces[fields[0]] = nsid
	}
	// Follow Linux' practise of identifying a leader task by its own leader
	// task set to zero.
	if TID == leader {
		leader = 0
	}
	return Task{
		TID:        TID,
		Leader:     leader,
		Namespaces: namespaces,
	}
}

// GetTaskIds returns the IDs of all tasks belonging to our process, as well as
// the task leader's task ID. If the information cannot be read (because the
// task leader has already exited), GetTaskIds return a nil list and a zero task
// leader ID.
func GetTaskIds() (taskids []uint32, taskleaderid uint32) {
	return getTaskIds("", uint32(os.Getpid()))
}

func getTaskIds(procroot string, PID uint32) (taskids []uint32, taskleaderid uint32) {
	procfsTaskDirItems, err := os.ReadDir(fmt.Sprintf(procroot+"/proc/%d/task", PID))
	if err != nil || len(procfsTaskDirItems) == 0 {
		return nil, 0
	}
	taskids = make([]uint32, 0, len(procfsTaskDirItems))
	for _, taskid := range procfsTaskDirItems {
		taskid, err := strconv.ParseUint(taskid.Name(), 10, 32)
		if err != nil || taskid == 0 || taskid > uint64(^uint32(0)) {
			continue
		}
		taskids = append(taskids, uint32(taskid))
	}
	return taskids, uint32(PID)
}
