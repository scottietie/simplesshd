# simplesshd

Purpose  
simplesshd is a minimal, self-contained SSH/SFTP server written in Go.  
It is a single binary—no extra keys, certificates, or config files required.  
Use it in experimental or transient environments (freshly-built hosts, embedded systems, CI sandboxes) when you need a quick, debug-friendly secure shell or SFTP access without installing a full OpenSSH stack.

Build  
1. go mod tidy  
2. go build  

The single binary that appears is all you need; copy it to the target and run.
