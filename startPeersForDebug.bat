start /wait cmd /c go build

for /L %%A IN (65001,1,65002) DO (
	start /MIN "Peer %%A" cmd /k  "blockchain.exe" -a 127.0.0.1:65000 -p %%A 
	
	timeout 2
)
