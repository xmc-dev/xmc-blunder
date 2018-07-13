# Gandalf

[![Build Status](https://travis-ci.org/xmc-dev/gandalf.svg?branch=master)](https://travis-ci.org/xmc-dev/gandalf)
[![Coverage Status](https://coveralls.io/repos/github/xmc-dev/gandalf/badge.svg)](https://coveralls.io/github/xmc-dev/gandalf)
[![GoDoc](https://godoc.org/github.com/xmc-dev/gandalf?status.svg)](https://godoc.org/github.com/xmc-dev/gandalf)

Gandalf is highly opinionated user permission library written in Go
specifically for the XMC project.

Gandalf's main purpose is to validate OAuth2 scopes according to a
*scope tree*. It's the program's job to check whether a user has access to a
resource or not.

## The Gandalf permission tree (the scope tree)

A scope represents a multitude of operations that the user can make on a
*resource*. For example, user `A` is allowed to *see* user `B`'s house, but
they're not allowed to *enter* it. To *see* and to *enter* are methods
specific to the house resource, more specifically user `B`'s house.

Each scope lives in a scope tree, meaning that all scopes have children and
parents. When a user has a scope they also have all of the scope's
children.

Example tree for a `user` object:

```
user
├── email
│   └── read
└── protected
```

* The `user` scope grants the user the permission do to anything they like
  to other `user` objects that can be accessed by them.

* The `user/email` scope grants the user the permission to read and write email
  addresses associated with the object.

* The `user/email/read` scope only lets the user to read the email addresses and
  nothing more.
