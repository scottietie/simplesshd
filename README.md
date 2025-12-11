# simplesshd

Purpose  
simplesshd is a minimal, self-contained SSH server written in Go.  
It is intended for experimental or transient environments—freshly-built hosts, embedded systems, CI sandboxes—where you need a quick, debug-friendly secure shell without installing a full OpenSSH stack.

Build  
1. go mod tidy  
2. go build  

The single binary that appears is all you need; copy it to the target and run.
