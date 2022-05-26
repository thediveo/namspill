/*

Package task discovers the Linux tasks (or, threads) of the current process.
Additionally, for each task the IDs of the attached namespaces are discovered.
This then allows checks to be made in order to detect namespaces "spilling" from
one goroutine to another due to incorrect OS-level thread locking.

Please note that the Go scheduler normally multiplexes goroutines onto threads
(tasks). However it is possible to lock a specific task to a single goroutine.

This package does not discover the relationship between tasks and goroutines;
this dynamically changes and is constantly controlled by the Go scheduler.

*/
package task
