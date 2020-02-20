package main

import (
	"bufio"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"unicode"
)

var (
	encryptedStrings []string
	inputFile        string
)

func init() {
	flag.StringVar(&inputFile, "input_file", "", "input file with the  obfsucated nanocore file")
}

func xor(enc []byte, key byte) (string, error) {
	ret := []byte{}

	for i := 0; i < len(enc); i++ {
		temp := enc[i] ^ key
		ret = append(ret, temp)
	}

	return string(ret), nil
}

func xorBrute(encodedStr []byte) (string, error) {
	switch string(encodedStr[0]) {
	case "0":
		// lazy
		return xor(encodedStr, 0)
	case "1":
		return xor(encodedStr, 1)
	case "2":
		return xor(encodedStr, 2)
	case "3":
		return xor(encodedStr, 3)
	case "4":
		return xor(encodedStr, 4)
	}

	return "", errors.New("not a valid nanocore encoding")
}

func file2lines(filePath string) ([]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return linesFromReader(f)
}

func linesFromReader(r io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func isASCII(s string) bool {
	for _, c := range s {
		if c > unicode.MaxASCII {
			return false
		}
	}
	return true
}

func main() {
	flag.Parse()

	var re = regexp.MustCompile(`(?m)"\b[0-9A-F]{2,}\b"`)
	if inputFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	lines, err := file2lines(inputFile)
	if err != nil {
		panic(err)
	}

	fileContent := ""

	for i, line := range lines {
		fileContent += line
		for _, match := range re.FindAllString(line, -1) {
			cleaned := strings.Replace(match, "\"", "", -1)
			dec, err := hex.DecodeString(cleaned)
			if err != nil {
				fileContent += "\n"
				continue
			}

			decodedStr, err := xorBrute(dec)
			if err != nil {
				fileContent += "\n"
				continue
			}

			if len(decodedStr) < 2 {
				fileContent += "\n"
				continue
			}

			if decodedStr[0:2] == "0x" {
				temp, err := hex.DecodeString(strings.Replace(decodedStr, "0x", "", -1))
				if err != nil {
					fileContent += "\n"
					continue
				}
				decodedStr = string(temp)
			}
			if isASCII(decodedStr) {
				fileContent += " ;" + decodedStr
				fmt.Printf("[+] decoded string at line %d: %s\n", i, decodedStr)
				fileContent += "\n"
			} else {
				fileContent += " ;" + "BINARYCONTENT"
				fileContent += "\n"
			}
		}
		fileContent += "\n"
	}

	ioutil.WriteFile("new_file.au3", []byte(fileContent), 0644)
}
