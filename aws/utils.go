package aws

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
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
