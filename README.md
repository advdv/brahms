# brahms
This is an experimental implementation of [Brahms: Byzantine resilient random membership sampling](https://www.cs.technion.ac.il/~gabik/publications/Brahms-COMNET.pdf). It describes a byzantine resilient protocol that creates a well-connected overlay network with each member only needing to knowing at most `O(∛n)` other peers.

## TODO
- [x] instead of Node ids, work with ip addresses and ports
- [ ] implement validation of the sample by probing
- [ ] decide, use and test an actual network transport
- [ ] fix the myriad of race conditions on shared memory variables
- [ ] implement a limited push with a small proof of work
- [ ] adjust l1 and l2 as the network grobs using an esimate as described [here](https://research.neustar.biz/2012/07/09/sketch-of-the-day-k-minimum-values/)
- [ ] fix send on closed channel bug with update calls with too short a context
- [ ] test if nodes can succesfully join by just pushing there ID
- [ ] use the crypto hash for node hashing also for sampling instead of farm hash
