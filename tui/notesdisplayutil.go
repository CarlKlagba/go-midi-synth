package main

import (
	"strconv"
	"strings"
)

var doremi = [12]string{
	"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B",
}

func DisplayNotes(notes []uint8) string {
	var sb strings.Builder
	for _, n := range notes {
		sb.WriteString(DisplayNote(n))
		sb.WriteString(" ")
	}
	return sb.String()
}

func DisplayNote(n uint8) string {
	u := int(n % 12)
	r := int(n/12) - 1
	var sb strings.Builder
	sb.WriteString(doremi[u])
	sb.WriteString(strconv.Itoa(r))
	return sb.String()
}
