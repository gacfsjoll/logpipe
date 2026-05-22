package router_test

import "bytes"

// This file exists solely to make the `bytes` import available to all test
// files in the router_test package without repeating the import in each file.
var _ = (*bytes.Buffer)(nil)
