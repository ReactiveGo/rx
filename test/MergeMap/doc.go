/*
MergeMap transforms the items emitted by an Observable by applying a
function to each item an returning an Observable. The stream of Observable
items is then merged into a single stream of items using the MergeAll
operator.

This operator was previously named FlatMap. The name FlatMap is deprecated as
MergeMap more accurately describes what the operator does with the observables
returned from the Map project function.
*/
package MergeMap
