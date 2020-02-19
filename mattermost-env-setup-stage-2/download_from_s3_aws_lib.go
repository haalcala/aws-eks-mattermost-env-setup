package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
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

func downloadFromS3(item string) error {
	fmt.Println("----- item:", item)

	folder := ""
	filename := ""

	if strings.LastIndex(item, "/") > 0 {
		folder = item[:strings.LastIndex(item, "/")]
		filename = item[strings.LastIndex(item, "/")+1:]
	} else {
		filename = item
	}

	fmt.Println("----- folder:", folder, "filename:", filename)

	if filename == "" {
		instances = instances - 1
		in_progress = in_progress - 1
		processed = processed + 1

		return nil
	}

	os.MkdirAll(path.Join("contents", folder), 0755)

	file, err := os.Create(path.Join("contents", item))

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
	in_progress = in_progress - 1
	processed = processed + 1

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
		_, err = reader.ReadString('\n')

		if err != nil {
			break
		}

		line_count = line_count + 1
	}

	count := 0

	for {
		line, err = reader.ReadString('\n')

		count = count + 1

		fmt.Printf("line #%d > Read %d characters\n", count, len(line))

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

	fmt.Println("Finished readling all lines.")

	return
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if info == nil || os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func displayProgress() {
	time.Sleep(1000)

	for processed < line_count {
		fmt.Println("------------------------------------------ processed:", processed, "in_progress:", in_progress)

		time.Sleep(3000)
	}
}

var instances = 0

var done = false
var processed = 0
var in_progress = 0
var line_count = 0

func main() {
	fmt.Println(os.Getenv("HOME"))
	if BUCKET == "" {
		fmt.Printf("Environment variable IMPORT_EXTERNAL_BUCKET not found.")
		os.Exit(1)
	}

	go displayProgress()

	err := readFileWithReadString("to-download-from-s3.txt", func(line string) {
		fmt.Println("Found line:", line)

		for instances > 10 {
			time.Sleep(1000)
		}

		if fileExists("contents/" + line) {
			return
		}

		instances = instances + 1
		in_progress = in_progress + 1

		go downloadFromS3(line)
	})

	if err != nil {
		fmt.Println("err:", err)

		os.Exit(1)
	}

	fmt.Println("Finished downloading files.")
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
