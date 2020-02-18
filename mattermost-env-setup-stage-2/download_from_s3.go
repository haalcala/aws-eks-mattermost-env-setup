package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

func Execute(command string, showCommand, showOutput bool) (error, string, string) {
	_command := strings.Split(command, " ")

	if showCommand {
		log.Println("--->>>:", command)
	}
	// fmt.Println("_command[0]:", _command[0])
	// fmt.Println("_command[1:]:", _command[1:])

	cmd := exec.Command(_command[0], _command[1:]...)
	// cmd.Stdin = strings.NewReader("some input")

	stdout, err := cmd.StdoutPipe()
	stderr, err := cmd.StderrPipe()

	if err != nil {
		log.Fatal(err)
	}

	cmd.Start()

	// cmd.Stdout = &stdout
	// cmd.Stderr = &stderr

	// err := cmd.Run()
	var output, errput string

	buf := bufio.NewReader(stdout) // Notice that this is not in a loop

	for {
		line, _, err := buf.ReadLine()

		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Println("Error while reading file, err", err)
			break
		}

		output = output + string(line) + "\n"

		// fmt.Println(string(line))
	}

	buf = bufio.NewReader(stderr) // Notice that this is not in a loop

	for {
		line, _, err := buf.ReadLine()

		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Println("Error while reading file, err", err)
			break
		}

		output = errput + string(line) + "\n"

		if showOutput {
			fmt.Println(string(line))
		}
	}

	return err, strings.Trim(output, "\n"), strings.Trim(errput, "\n")
}

func downloadFromS3(file string) error {
	err, ret, stderr := Execute(fmt.Sprintf("aws s3 cp s3://%s/%s contents/%s --profile prod", os.Getenv("__IMPORT_EXTERNAL_BUCKET__"), file, file), true, true)

	if err != nil {
		fmt.Println("err:", stderr)

		os.Exit(1)
	}

	fmt.Println("ret:", ret)

	instances = instances - 1

	return err
}

func readFileWithReadString(fn string, handleLine func(line string)) (err error) {
	fmt.Println("readFileWithReadString")

	file, err := os.Open(fn)
	defer file.Close()

	if err != nil {
		return err
	}

	// Start reading from the file with a reader.
	reader := bufio.NewReader(file)

	var line string

	maxLines := 0

	for {
		line, err = reader.ReadString('\n')

		fmt.Printf(" > Read %d characters\n", len(line))

		// Process the line here.
		// fmt.Println(" > > " + limitLength(line, 50))

		handleLine(strings.Trim(line, "\n"))

		maxLines = maxLines + 1

		if err != nil || maxLines >= 200 {
			break
		}
	}

	if err != io.EOF {
		fmt.Printf(" > Failed!: %v\n", err)
	}

	return
}

var instances int = 0

func main() {

	err := readFileWithReadString("to-download-from-s3.txt", func(line string) {
		fmt.Println("Found line:", line)

		for instances > 10 {
			time.Sleep(1000)
		}

		instances = instances + 1

		go downloadFromS3(line)

	})

	if err != nil {
		fmt.Println("err:", err)

		os.Exit(1)
	}
}
