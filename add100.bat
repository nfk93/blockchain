for /L %%A IN (65021,1,65100) DO (
	start /MIN "Peer %%A" cmd /k  "blockchain.exe" -a 127.0.0.1:65000 -p %%A 
	
	timeout 2
)
