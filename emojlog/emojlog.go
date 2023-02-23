package emojlog

import (
	"fmt"
	"time"
)

func PrintLogMessage(value string, msgType string) {
	var result string
	switch msgType {
	case "info":
		result = generateInfo(value)
	case "changed":
		result = generateChanged(value)
	case "debug":
		result = generateDebug(value)
	case "warning":
		result = generateWarning(value)
	case "error":
		result = generateError(value)
	default:
		result = ""
	}
	fmt.Println(result)
}

func generateInfo(value string) string {
	initialValue := " 🟢 INFO:    🕔 " + generateTime() + ": 📄 "
	return initialValue + value
}

func generateChanged(value string) string {
	initialValue := " 🔶 CHANGED: 🕔 " + generateTime() + ": 📄 "
	return initialValue + value
}

func generateDebug(value string) string {
	initialValue := " 🔷 DEBUG:   🕔 " + generateTime() + ": 📄 "
	return initialValue + value
}

func generateWarning(value string) string {
	initialValue := " 🔴 WARNING: 🕔 " + generateTime() + ": 📄 "
	return initialValue + value
}

func generateError(value string) string {
	initialValue := " 🚫 ERROR:   🕔 " + generateTime() + ": 📄 "
	return initialValue + value
}

func generateTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
