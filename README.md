# Toy Stack Machine in Go

An experiment grown out of repeated solutions to [a toy integer
search/programming problem][intsearch]:
- ... where-in a [C stack machine][intcstack] solved it
- ... where-in a [Go register-architecture VM][intgoreg] solved it

The key twist is that the vm has a `FORK` instruction:
- `FORK` works like `JUMP`, except both paths are taken
- this works by copying the current machine state and:
  - having the copy `JUMP` normally
  - while the original continues on, as if it ignored the `FORK`
- as a corollary, a `BRANCH` is like `FORK`, but the original machine makes the
  `JUMP` while the copy continues

Conditional forms of `FORK` and `BRANCH` work similarly to `JUMP` wrt `JZ` and
`JNZ`:
- `FZ`/`FNZ` are fork if (non-)zero
- `BZ`/`BNZ` are branch if (non-)zero

Of course this means that we need some way of handling multiple descendant
copies while running a machine. Perhaps the simplest thing to do:
- push copies onto a queue of pending machines
- after the current machine run ends, pop and run a machine from the queue
- continue like this until the queue is empty, or an abnormal termination occurs

# Status

- the api is missing tracing support, this will likely be the next
  thing to come as it greatly aids writing tests
- many basic ops are coded, but not yet tested
- fork/branch infrastructure is ~80% in place, need to complete the
  `context` interface
- better support for normal termination "errors" is planned:
  symbolication will be attempted for non-zero halt codes, probably
  through a lookup provided by `context`
- `CALL`/`RET` are currently unimplemented due to exposing dissonance
  between
  memory and stack granularity; it's probably time to change the stack units to be "words" rather than "bytes"
  "word"-sized
- adding gather/scatter ops seems like a good idea at this point
  (noted in code at two points); additionally gather may provide a compelling
  final answer to "how/what result value(s) do we extract when finished?"

**All Code** is currently on the [`dev`][dev] branch.

Current plan is to get the vm itself into a moderately solid state, and then to
start building a small [FORTH][forth]-like language on top of it.
architectural trade-offs:

[intsearch]: https://github.com/jcorbin/intsearch
[intcstack]: https://github.com/jcorbin/intsearch/tree/c_stack_machine_2015-11
[intgoreg]: https://github.com/jcorbin/intsearch/tree/go_2016-04
[dev]: (https://github.com/jcorbin/intsearch/tree/dev)
[forth]: https://en.wikipedia.org/wiki/Forth_(programming_language)
