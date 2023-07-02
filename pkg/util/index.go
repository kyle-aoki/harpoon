package util

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

func Check(err error) {
	if err != nil {
		panic(err)
	}
}

func Must[T any](t T, err error) T {
	Check(err)
	return t
}

var F = fmt.Sprintf

func Bash(command string) (string, error) {
	log.Println("executing:", command)
	cmd := exec.Command("bash", "-c", command)
	stdout := Must(cmd.StdoutPipe())
	stderr := Must(cmd.StderrPipe())
	err := cmd.Start()
	if err != nil {
		return "", err
	}
	sout := cmdText(stdout)
	serr := cmdText(stderr)
	if len(serr) != 0 {
		return sout, errors.New(serr)
	}
	return sout, nil
}

func cmdText(txt io.ReadCloser) string {
	if txt == nil {
		return ""
	}
	scanner := bufio.NewScanner(txt)
	var strs []string
	for scanner.Scan() {
		strs = append(strs, scanner.Text())
	}
	return strings.Join(strs, "\n")
}

func JsonLinesToSlice[T any](jsonLines string) []*T {
	var ts []*T
	if len(jsonLines) == 0 {
		return ts
	}
	for _, line := range strings.Split(jsonLines, "\n") {
		var t T
		Check(json.Unmarshal([]byte(line), &t))
		ts = append(ts, &t)
	}
	return ts
}

func IfOneTrue(bools ...bool) bool {
	for i := 0; i < len(bools); i++ {
		if bools[i] {
			return true
		}
	}
	return false
}

func Contains(arr1 []int, target int) bool {
	for i := 0; i < len(arr1); i++ {
		if arr1[i] == target {
			return true
		}
	}
	return false
}

func ToStr(i uint16) string {
	return strconv.FormatInt(int64(i), 10)
}
