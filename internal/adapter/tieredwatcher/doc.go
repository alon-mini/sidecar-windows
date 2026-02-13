// Package tieredwatcher implements tiered file watching with a HOT tier using
// real-time fsnotify and a COLD tier using periodic polling, reducing file
// descriptor usage while keeping recently active sessions responsive.
package tieredwatcher
