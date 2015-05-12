[![Build Status](https://travis-ci.org/lambdasoup/go-netlink.svg?branch=master)](https://travis-ci.org/lambdasoup/go-netlink)

go-netlink
==========

This is a suite of Go packages to interface with the Linux netlink subsystem.

go-netlink was created to talk to a Maxim iButton temperature sensor via netlink. The provided suite so far does not do much more than to do exactly that. Netlink / Connector / One-Wire have been separated into hierarchical layers to reflect the kernel's architecture. Feel free to extend this to more Netlink / Connector or One-Wire endpoints.

The project's license is GPL3+

# ibutton tool installation

Go package management can install ibutton directly from github:
```
go install github.com/lambdasoup/go-netlink/ibutton/ibutton
```

# ibutton tool usage

start a new mission
```
ibutton -command start
```

stop the currently running mission
```
ibutton -command stop
```

print out the sample log
```
ibutton -command read
```

show the button status
```
ibutton -command status
```

clear the button mission memory
```
ibutton -command clear
```
