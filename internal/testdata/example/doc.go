// Package example is an example of how Luno core repo can use weld to generate state.
// It also acts as test input.
//
// Note that the go.mod and go.sum files have been renamed to prevent go.sum sporadically changing.
// Since the go.mod file uses a "replace" directive to find the bitx module, the
// bitx pinned version is ignored. This resulted in sporadic updates to go.sum whenever
// bitx imported new libraries. To workaround this issue, we keep static renamed versions of
// of go.mod and go.sum checked into git. gen_test#TestMain renames them correctly, and then ensures
// any changes are reverted afterwards.
package example
