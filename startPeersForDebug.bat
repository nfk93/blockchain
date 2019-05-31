del blockchain.exe
go build
start "Network Owner" cmd /k blockchain.exe -log=true -slot_duration=1 -hardness=0.4 -p=12000
timeout 5

RD /S /Q out
MD out
for /L %%A IN (65001,1,65030) DO (
	start /MIN "Peer %%A" cmd /k  "blockchain.exe" -a 127.0.0.1:12000 -p %%A

	timeout 2
)
