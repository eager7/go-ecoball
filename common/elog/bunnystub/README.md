# BunnyStub

BunnyStub is a common components library for logbunny

## Why we need BunnyStub

logbunny is a flaxable logger, you can construct your with different IO writer, Here we made some IO writer for logbunny user convenient using.

## What's in it
* time routated IO writer
* volume routated IO writer

test case | testing times | cost on per-operation | allocation on per-option | allocation times on per-option
----------|---------------|-----------------------|--------------------------|-------------------------------
BenchmarkWrite-4           |     10000000         |      472 ns/op           |   36 B/op      |    1 allocs/op
BenchmarkParallelWrite-4   |     10000000         |      639 ns/op           |   36 B/op      |    1 allocs/op

## Contribute && TODO
Any new feature inneed pls [put up an issue](https://gitlab.meitu.com/gocommons/bunnystub/issues/new)
