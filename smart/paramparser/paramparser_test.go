package paramparser

import (
	"bufio"
	"fmt"
	"os"
	"testing"
)

func TestParseParams1(t *testing.T) {
	file, err := os.Open("testfile_noerror")
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	defer file.Close()

	reader := bufio.NewScanner(file)
	for reader.Scan() {
		params, err := ParseParams(reader.Text())
		if err != nil {
			t.Errorf(err.Error())
			fmt.Println("Error in: " + reader.Text())
			fmt.Println(err.Error())
			fmt.Println("")
		} else {
			fmt.Println(params)
			fmt.Println("")
		}
	}
}

func TestParseParams2(t *testing.T) {
	// all these tests should have errors
	file, err := os.Open("testfile_error")
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	defer file.Close()

	reader := bufio.NewScanner(file)
	for reader.Scan() {
		params, err := ParseParams(reader.Text())
		if err != nil {
			fmt.Println("Error in: " + reader.Text())
			fmt.Println(err.Error())
			fmt.Println("")
		} else {
			t.Errorf("no error in testcase: %s\nresulting value: %s", reader.Text(), params)
		}
	}
}
