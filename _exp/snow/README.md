# Wavelet

Wavelet: https://medium.com/perlin-network/wavelet-beta-df9c738747e2 And the
whitepaper: https://wavelet.perlin.net/whitepaper.pdf Based on the original
snowball algorithm: https://ipfs.io/ipfs/QmUy4jh5mGNZvLkjies1RWM4YuvJh5o2FYopNPVYwrRVGV
The novel aspect is the finalization through peer querying

# Coordicity

IOTA got a new iteration: https://blog.iota.org/coordicide-e039fd43a871 described
in the whitepaper here: https://files.iota.org/papers/Coordicide_WP.pdf
Consensus combines "Fast probabilistic consensus" described by IOTA cofounder
and "Cellular consensus"

# NKN

Has been using cellular consensu from the beginning, with nice visualizations
over here: https://medium.com/nknetwork/the-origin-of-nkn-scalable-moca-consensus-based-on-cellular-automata-20d9985130c0

# IDEA:


Given a certain proof, any member can initiate a cellular consensus across the whole network. Only one can be started at a time.

 - But it is still probabilistic consensus, why is that any better then a chain? Or put another way, doesn't it just
   proof that whatever nr of samples are performed that they all are on this track. Sampling is always a subset
-  What about eclipse attacks, specifically on kademlia, what if a node gets isolated? It can be tricked into
   thinking there has been consensus on a certain block?
-  What happens to the dag when snowball is running, is there a rule that makes them all build on the critical
   transaction? or stop the world?
-  what if a majority of the network is behind when snowball is running? It is synchronous so everyone needs to
   be a the same spot?


### The Parts

- Gossip: to exchange blocks with N transactions to form a...
- Graph: where each vertex is a block, referencing M parents. Its structure should allow for...
- Tip-Selection: such that new blocks can be proposed and broadcasted using the gossip protocol. This gives us...
- Probabilistic consensus: some economic certainty on the transaction order. Once in a while this requires....
- Finalization: where every one synchronously/interactively decides that a part of the graph is final.  

The graph is kept in memory, if it is possible to use a finalization mechanism to snapshot the state

### Finalization

Finalization of a round (or generation) means that it is impossible for any late
arriving block (however perfect) to change the order of the data in that round.
This means that adding blocks should give it more value whenever new blocks arrive.

synchronous finalization: You had to be there to believe it, we agreed.
asynchronous finalization: if you look at the structure i think you would agree. But what if not everyone has information encoded into the chain to prove this.
