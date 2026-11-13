package main

import (
	"crypto/rand"
	"fmt"
)

func generateUUID() (string, error) {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		return "", err
	}

	// Set RFC 4122 version 4 bits: xxxxxxxx-xxxx-4xxx-xxxx-xxxxxxxxxxxx
	id[6] = (id[6] & 0x0f) | 0x40

	// Set RFC 4122 variant bits: xxxxxxxx-xxxx-xxxx-8xxx-xxxxxxxxxxxx
	id[8] = (id[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		id[0:4], id[4:6], id[6:8], id[8:10], id[10:]), nil
}

func main() {
	uuid, err := generateUUID()
	if err != nil {
		panic(err)
	}

	fmt.Println(uuid)
}
