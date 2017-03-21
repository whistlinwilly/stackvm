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
- time to bite the bullet and use the unsafe package:
  - need a `(*page).ref(addr uint32) (p *uint32)`
  - use it anywhere we have a pop/push pattern
  - or a addr/store pattern
  - e.g. binary op accumulator and swap
- we could enforce a stricter memory model if:
  - `(*Mach).Load` took an additional "max memory size"...
  - ...then it loads the program at `cbp + maxMem`...
  - ...and stores the end of the program section as a limit...
  - ...that fetch and store check against, causing a segfault.
  - The biggest reason that I'm not currently doing this, is that it'd
    make it difficult to impossible to implement a dynamic compiler,
    like the planned forth experiment.
  - However if instead of going towards forth, all of the
    "compilation" or "assembly" happens externall to the machine, e.g.
    in Go, then this stricter memory model becomes highly desirable as
    it helps to catch programming errors.
- should provide some sort of static program verification; at least
  "can I decode it straight thru?"

**All Code** is currently on the [`dev`][dev] branch.

Current plan is to get the vm itself into a moderately solid state, and then to
start building a small [FORTH][forth]-like language on top of it.
architectural trade-offs:

[intsearch]: https://github.com/jcorbin/intsearch
[intcstack]: https://github.com/jcorbin/intsearch/tree/c_stack_machine_2015-11
[intgoreg]: https://github.com/jcorbin/intsearch/tree/go_2016-04
[dev]: (https://github.com/jcorbin/intsearch/tree/dev)
[forth]: https://en.wikipedia.org/wiki/Forth_(programming_language)
