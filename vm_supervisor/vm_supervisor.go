package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

func main() {
	// Get env vars passed from "hoster vm start"
	vmStartCommand := os.Getenv("VM_START")
	vmName := os.Getenv("VM_NAME")
	logFileLocation := os.Getenv("LOG_FILE")

	// Set the process name
	procName := "vm supervisor: " + vmName
	os.Args[0] = procName

	// Create or open the log file for writing
	logFile, err := os.OpenFile(logFileLocation, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0640)
	if err != nil {
		log.Fatal("Unable to open log file: " + err.Error())
	}
	// Redirect the output of log.Fatal to the log file
	log.SetOutput(logFile)

	// Start the process
	parts := strings.Fields(vmStartCommand)
	for {
		cmd := exec.Command(parts[0], parts[1:]...)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatalf("Failed to create stdout pipe: %v", err)
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			log.Fatalf("Failed to create stderr pipe: %v", err)
		}

		var wg sync.WaitGroup
		wg.Add(2)

		stdoutReader := bufio.NewReader(stdout)
		go func() {
			defer wg.Done()
			readAndLogOutput(stdoutReader, "stdout")
		}()

		stderrReader := bufio.NewReader(stderr)
		go func() {
			defer wg.Done()
			readAndLogOutput(stderrReader, "stderr")
		}()

		done := make(chan error)
		startCommand(cmd, done)

		wg.Wait()

		if err := <-done; err != nil {
			log.Printf("Command failed: %v", err)
			if exitError, ok := err.(*exec.ExitError); ok {
				if status, ok := exitError.Sys().(interface{ ExitStatus() int }); ok {
					exitCode := status.ExitStatus()
					if exitCode != 1 {
						log.Printf("Command returned non-zero exit code: %d, restarting...", exitCode)
						continue
					}
				}
			}
			log.Fatal("Failed to get exit code")
		}

		time.Sleep(time.Second)
	}
}

func readAndLogOutput(reader *bufio.Reader, name string) {
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed to read %s: %v", name, err)
		}
		line = strings.TrimSpace(line)
		if line != "" {
			log.Printf("[%s] %s\n", name, line)
		}
	}
}

func startCommand(cmd *exec.Cmd, done chan error) {
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start command: %v", err)
	}
	go func() {
		done <- cmd.Wait()
	}()
}
