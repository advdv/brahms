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
- [x] decide, use and test an actual network transport
- [x] implement and test the agent
- [x] create and test a command that runs the agent as a process
- [x] (fix) bug that prevent the network to grow from 1 to 2
- [x] test if nodes can succesfully join by just pushing there ID
- [x] (fix) leave of node causes peers to keep probing on real agent proc
- [x] make validation and update timeouts configurable
- [x] (fix) remove core.IsActive lock contention
- [x] (fix) remove readView lock contention
- [x] (fix) pull easily runs into deadline exceeded
- [x] (fix) prevent pulling of nodes that were recently invalidated
- [x] refactor probing to not probe double samples
- [x] (fix) make sure its not possible for others to push others node's "self" info
- [x] finish basic agent implementation
- [x] add a general interface to the agent to dissemate messages
- [ ] add a simple way to dissemate custom message in brahmsd

- [ ] add a cellular consensus mechanism on sets
- [ ] probe only a part of the sampled nodes at a time (round-robin, like SWIM)

- [ ] implement a limited push with a small proof of work
- [ ] adjust l1 and l2 as the network grobs using an esimate as described [here](https://research.neustar.biz/2012/07/09/sketch-of-the-day-k-minimum-values/)
- [ ] use the crypto hash for node hashing also for sampling instead of farm hash
- [ ] store the node's sample on disk
- [ ] measure if lock contention on sampler is too high
- [ ] only full shutdown gossip agent if no messages arrive anymore
- [ ] add indirect probing if direct probing doesn't work (similar to SWIM)

## When to refresh the view
Refreshing the view seems to be an open design decision.
- Always refresh unless push is too big
- Only refresh if there are pushed or pullsed nodes
- Only refresh if there are both push and pulls

Always refresh makes sure it is the most up-to-date but causes the view to
easily fluctuate. but thats were the history sample is for?
