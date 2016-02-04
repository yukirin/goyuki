package main

// Name is executable name of this application
const Name string = "goyuki"

// Version is version string of this application
const Version string = "0.1.0"

// GitCommit describes latest commit hash.
// This value is extracted by git command when building.
// To set this from outside, use go build -ldflags "-X main.GitCommit \"$(COMMIT)\""
var GitCommit string
