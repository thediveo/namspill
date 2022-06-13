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

import "github.com/thediveo/namspill/task"

// Task represents an individual Linux task (thread) of the running Go process;
// it is a type alias for task.Task.
type Task = task.Task

// Tasks returns information about the tasks currently belonging to this
// process. If the tasks cannot properly be determined, Tasks returns nil
// instead.
//
// This is a re-export of the task.Tasks discovery function for convenience.
func Tasks() []Task {
	return task.Tasks()
}
