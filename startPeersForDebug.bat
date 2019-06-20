del blockchain.exe
go build
start "Network Owner" cmd /k blockchain.exe -run_locally -slot_duration=5 -hardness=0.5 -epoch_length=20 finalize_gap=100
timeout 5

RD /S /Q out
MD out
for /L %%A IN (65001,1,65005) DO (
	start /MIN "Peer %%A" cmd /k  "blockchain.exe" -a 127.0.0.1:65000 -p %%A

	timeout 2
)
