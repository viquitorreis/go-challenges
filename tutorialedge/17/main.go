package main

import (
	"encoding/base64"
	"fmt"
	"strings"
)

func main() {
	fmt.Println("Decode the Secret")

	message := "VEZEU0ZVVFVTSk9I"
	result := DecodeSecret(message)
	fmt.Println(result)

}

func DecodeSecret(message string) string {
	b64, _ := base64.StdEncoding.DecodeString(message)
	return string(alphaBeticalShift(b64))
}

func alphaBeticalShift(s []byte) string {
	var st strings.Builder
	for _, ch := range s {
		char := ch - 1
		st.WriteByte(char)
	}
	return st.String()
}
