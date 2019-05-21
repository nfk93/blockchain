package paramparser

import (
	"bufio"
	"fmt"
	"os"
	"testing"
)

func TestParseParams(t *testing.T) {
	file, err := os.Open("testfile")
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	defer file.Close()

	reader := bufio.NewScanner(file)
	for reader.Scan() {
		params, err := ParseParams(reader.Text())
		if err != nil {
			fmt.Println(reader.Text())
			t.Errorf(err.Error())
			return
		}
		fmt.Println(params)
	}
}
