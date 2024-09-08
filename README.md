# Refactor on UCAN Tokens in Go by KenCloud Technologies
![UCAN](https://img.shields.io/badge/UCAN-v0.7.0-blue)

Originally by @b5 as one of the first from scratch implementations of UCAN outside of the Fission teams initial work in TypeScript / Haskell.
ucan-wg/go-ucan: User-Controlled Authorization Network (UCAN) tokens in go
https://github.com/ucan-wg/go-ucan

Now by @sssion as one of the iteration on @b5's implementations of UCAN.

**If you're interested in updating this codebase to the 1.0 version of the UCAN spec, [get involved in the discussion »](https://github.com/orgs/ucan-wg/discussions/163)**

## About UCAN Tokens

User Controlled Authorization Networks (UCANs) are a way of doing authorization where users are fully in control. OAuth is designed for a centralized world, UCAN is the distributed user controlled version.

### UCAN Gopher

![](https://ipfs.runfission.com/ipfs/QmRFXjMjVNwnYki8jGwFBh3zcY5m7zo5oAcNoyS1PSgzAY/ucan-gopher.png)

Artwork by [Bruno Monts](https://www.instagram.com/bruno_monts). Thank you [Renee French](http://reneefrench.blogspot.com/) for creating the [Go Gopher](https://blog.golang.org/gopher)

# GO-UCAN-KC
## Overview
This Go project implements User-Controlled Authorization Network (UCAN) tokens, providing a decentralized authorization mechanism. Originally based on work by KenCloud Technologies, this project has been refactored and extended to meet the requirements of modern, distributed systems.

### Key Features
- Decentralized Authorization: Users maintain full control over their authorization tokens.
- Flexible Token Building: Easily create and customize UCAN tokens with various capabilities and constraints.
- Chain of Trust: Supports token chaining to enable attestations and delegations.
- Comprehensive Testing: Includes a suite of tests to ensure reliability and functionality.
### Installation
To get started with the Go UCAN project, clone the repository and install the necessary dependencies:

``` bash
git clone https://github.com/yourusername/go-ucan-project.git
cd go-ucan-project
go mod tidy
```

### Usage
Here's a basic example of how to use the UCAN builder to create a token:

```go
package main

import (
	"github.com/KenCloud-Tech/go-ucan-kc"
	"log"
)

func main() {
	builder := go_ucan_kl.DefaultBuilder().
		IssuedBy(issuerKey).
		ForAudience(audienceDid).
		WithLifetime(3600) // 1 hour

	token, err := builder.Build()
	if err != nil {
		log.Fatalf("Failed to build token: %v", err)
	}

	log.Println("Token:", token)
}
```