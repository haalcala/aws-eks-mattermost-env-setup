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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var BUCKET string = os.Args[1]
var REGION string = os.Args[2]

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

func downloadFromS3(item string) error {
	fmt.Println("----- item:", item)
	fmt.Println("-----", item[:strings.LastIndex(item, "/")])

	filename := item[strings.LastIndex(item, "/")+1:]

	fmt.Println("filename:", filename)

	if filename == "" {
		return nil
	}

	os.MkdirAll("contents/"+item[:(strings.LastIndex(item, "/"))], 0755)

	file, err := os.Create("contents/" + item)

	if err != nil {
		exitErrorf(">>> Unable to open file %q, %v", item, err)
	}

	defer file.Close()

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(REGION), Credentials: credentials.NewSharedCredentials("", "prod")},
	)

	if err != nil {
		exitErrorf("Unable to create session, %v", err)
	}
	_, err = sess.Config.Credentials.Get()

	if err != nil {
		exitErrorf("Unable to verify credentials, %v", err)
	}

	downloader := s3manager.NewDownloader(sess)

	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(BUCKET),
			Key:    aws.String(item),
		})

	if err != nil {
		fmt.Println("err:", err)
		// exitErrorf("Unable to download item %q, %v", item, err)
	}

	fmt.Println("Downloaded", file.Name(), numBytes, "bytes")

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

	for {
		line, err = reader.ReadString('\n')

		fmt.Printf(" > Read %d characters\n", len(line))

		// Process the line here.
		// fmt.Println(" > > " + limitLength(line, 50))

		handleLine(strings.Trim(line, "\n"))

		if err != nil {
			break
		}
	}

	if err != io.EOF {
		fmt.Printf(" > Failed!: %v\n", err)
	}

	return
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if info == nil || os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

var instances int = 0

func main() {
	fmt.Println(os.Getenv("HOME"))
	if BUCKET == "" {
		fmt.Printf("Environment variable IMPORT_EXTERNAL_BUCKET not found.")
		os.Exit(1)
	}

	err := readFileWithReadString("to-download-from-s3.txt", func(line string) {
		fmt.Println("Found line:", line)

		for instances > 10 {
			time.Sleep(1000)
		}

		if fileExists("contents/" + line) {
			return
		}

		instances = instances + 1

		go downloadFromS3(line)
	})

	if err != nil {
		fmt.Println("err:", err)

		os.Exit(1)
	}
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
