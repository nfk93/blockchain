# Proof-of-Stake Blockchain

This is a CLI implementing a synchronous of a Proof-of-Stake blockchain protocol.

### Installing

1. Download and setup GO from http://golang.org/
2. Setup your GOPATH environment variable
3. Run "go get https://github.com/nfk93/blockchain"
4. Run "go build" in the root folder of the project, this produces blockchain.exe

Alternatively, simply clone the the project to %GOPATH%/src/github.com/nfk93/blockchain and build it there

### Running

Running "blockchain.exe -help" will display the following:

```
Usage of blockchain.exe:
  -a string
        Address to connect to, INCLUDING PORT, (if not set, start own network)
  -epoch_length uint
        Specify the epoch length, only set this is if you're starting a new network (default 100)
  -finalize_gap uint
        Specify the finalization gap, only set this is if you're starting a new network (default 1500)
  -hardness float
        Specify hardness (default 0.1)
  -log
        will save simple log of all blocks in dot format if set (default false)
  -p string
        Port to be listening on used for p2p (default "65000")
  -slot_duration int
        Specify the slot length in seconds (default 10)
```

To start a new blockchain run
```
blockchain.exe -p=12345 -slot_duration=15 -epoch_length=20 -finalize_gap=100 -hardness=0.5
```
replacing 12345 with the port of your desire.

Before starting the blockchain protocol, you will want to connect all other participants by running 
```
blockchain.exe -a=<some_address> -p=12346
```
where <some_address> must be the address of someone already on the network, including his port 
(for example, if the above case was run on network, a good choice would be "127.0.0.1:12345")

When everyone is connected, the one who started the network can type start into his CLI, at which point everyone on the network should report "Genesis received! Starting blockchain protocol"
