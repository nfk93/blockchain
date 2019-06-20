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
        Set to write log of tree in each slot to /out (default false)
  -p string
        Port to be listening on used for p2p (default "65000")
  -run_locally
        Set if only running locally (default true) (default true)
  -slot_duration int
        Specify the slot length in seconds (default 10)
```

To start a new blockchain run
```
blockchain.exe -p=65000 -slot_duration=5 -epoch_length=20 -finalize_gap=100 -hardness=0.5
```
replacing the above example values with your desire

Before starting the blockchain protocol, you will want to connect all other participants by running 
```
blockchain.exe -a=<some_address> -p=65001
```
where <some_address> must be the address of someone already on the network, including his port 
(for example, if the above case was run, a good choice would be "127.0.0.1:65000", since it runs locally unless setting the flag -run_locally=false)

When everyone is connected, the one who started the network can type start into his CLI, at which point everyone on the network should report "Genesis received! Starting blockchain protocol"

### Running a test example

To run a test example, you can use the bat file startPeersForDebug.bat, which starts a network with the parameters used in the example above and connects 5 peers to it. When the peers are connected (takes roughly 15 seconds), use "start" in the commandline window titled "Network Owner". Each peer should now look like this:
```
my ip is: 127.0.0.1
CONNECTING TO EXISTING NETWORK AT  127.0.0.1:65000
This client is ready for the Blockchain protocol to start. Use -h or --help for further commands!
2019/06/20 11:04:35 Genesis received! Starting blockchain protocol
»
```
To test that the blockchain is indeed working, enable verbose on some of the peers by typing 'verbose' in the commandline. Use the two debug methods "debug-autotrans" and "debug-trans5" to make some test transactions. The full list of commands can be found by typing -h or --help, which lists 
```
  Command:                            Description:

  exit                                Exit this program

  start                               Begins the blockchain protocol

  verbose                             Switches verbose mode between on and off, initially off

  id                                  Show your public key in short and which port you are listening on

  network                             Print out the network list of who you are connected to

  publicKeys                          Print list of know Public keys in network

  ledger                              Print the current ledger

  final                               Print the last finalized ledger

  seenTrans                           Print list of seen transactions

  contracts                           Prints a list of all currently active contracts

  contractInfo ADDRESS                Prints info of a given contract
                                      ADDRESS: The address of a given contract

        -o <string>                   Path with filename of output file. Path should be without file extension.

  transaction RECEIVER AMOUNT         Send Amount to the Receiver
                                      RECEIVER: A 10 digit prefix of senders publicKey hash
                                      AMOUNT: Positive integer of amount to transfer

  call ADDRESS GAS                    Makes a contract call to the specified contract
                                      ADDRESS: The address of the contract
                                      GAS: Positive integer of how much gas to include

        -entry <string>               Default: main. Used to specify the entry in the contract to call
        -amount <uint>                Default: 0. Non negative integer of amount included in a contract call
        -params <string>              Default: "()". Used to specify parameters to include in contract call

  init CODE GAS PREPAID STORAGE
                                      CODE: path to code file
                                      GAS: Positive integer of how much gas to include
                                      PREPAID: Positive integer of how much prepaid money to attached at contract
                                      STORAGE: Positive integer of max storage usage for a contract

  debug-trans5                        Sends 1/20 of your stake to 5 random users in the network

  debug-autotrans                     Toggle: Sends 1/50 of your stake to 5 random users in the network every 5 seconds
```
