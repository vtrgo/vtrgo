package main

import (
	"log"
)

func welcome(name string) string {
	if name != "" {
		log.Printf("Hello, %s", name)
		return name
	}
	log.Printf("No text")
	return "Nothing"
}
