start cmd /c go build
timeout 7
start cmd /k blockchain.exe
timeout 3
start cmd /k blockchain.exe -a 127.0.0.1:65000 -p 65001
start cmd /k blockchain.exe -a 127.0.0.1:65000 -p 65002
start cmd /k blockchain.exe -a 127.0.0.1:65000 -p 65003
start cmd /k blockchain.exe -a 127.0.0.1:65000 -p 65004
start cmd /k blockchain.exe -a 127.0.0.1:65000 -p 65005
start cmd /k blockchain.exe -a 127.0.0.1:65000 -p 65006
