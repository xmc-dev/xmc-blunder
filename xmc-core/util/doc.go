// Package util provides utility functions for managing certain XMC objects.
// Most of these functions are supposed to operate in a Datastore transaction group,
// yet the actual management of this group (BeginGroup, Rollback, Commit) should not be done
// in the body of the function, but in the body of the calling function.
package util
