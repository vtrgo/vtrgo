package main

import (
	"log"
)

func welcome(s string) string {

	st := s
	if st != "" {
		log.Printf("Hello, %s", st)
		return st
	}
	log.Printf("No text")
	return "Nothing"
}
