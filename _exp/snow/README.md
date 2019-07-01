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

A sync with encryption to "join" a certain cellular consensus exchange across certain choices. like joining a chat room.


### The Parts

- Gossip: to exchange blocks with N transactions to form a...
- Graph: where each vertex is a block, referencing M parents. Its sture should allow for...
- Tip-Selection: such that new blocks can be proposed and broadcasted using the gossip protocol. This gives us...
- Probabilistic consensus: some economic certainty on the transaction order. Once in a while this requires....
- Finalization: where every one synchronously/interactively decides that a part of the graph is final.  

The graph is kept in memory, if it is possible to use a finalization mechanism to snapshot the state
