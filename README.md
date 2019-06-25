# brahms
This is an experimental implementation of [Brahms: Byzantine resilient random membership sampling](https://www.cs.technion.ac.il/~gabik/publications/Brahms-COMNET.pdf). It describes a byzantine resilient protocol that creates a well-connected overlay network with each member only needing to knowing at most `O(∛n)` other peers.

## TODO
- [x] instead of Node ids, work with ip addresses and ports
- [x] fix send on closed channel bug with update calls with too short a context
- [x] implement validation of the sample by probing
- [x] move transports to sub-package
- [x] refactor core
  - [x] reduce nr of public methods necessary for tests
  - [x] formalize alive state
  - [x] fix the myriad of race conditions on shared memory variables
  - [x] test for concurrent access
- [x] fix bug that causes cores sample not to be stable over time
- [x] fix big that cause some deactivated cores to linger in the sample of others
- [ ] decide, use and test an actual network transport
- [ ] implement a limited push with a small proof of work
- [ ] test if nodes can succesfully join by just pushing there ID
- [ ] adjust l1 and l2 as the network grobs using an esimate as described [here](https://research.neustar.biz/2012/07/09/sketch-of-the-day-k-minimum-values/)
- [ ] use the crypto hash for node hashing also for sampling instead of farm hash
- [ ] probe only a part of the sampled nodes at a time
