package tui

import "testing"

type testParam struct {
	n        uint8
	expected string
}

var testParams = []testParam{
	{0, "C-1"},
	{12, "C0"},
	{24, "C1"},
	{36, "C2"},
	{48, "C3"},
	{60, "C4"},
	{72, "C5"},
	{84, "C6"},
	{1, "C#-1"},
	{61, "C#4"},
	{73, "C#5"},
	{2, "D-1"},
	{3, "D#-1"},
	{4, "E-1"},
	{5, "F-1"},
	{6, "F#-1"},
	{7, "G-1"},
	{8, "G#-1"},
	{9, "A-1"},
	{10, "A#-1"},
	{23, "B0"},
}

func TestConvertIntToNote(t *testing.T) {

	for _, test := range testParams {
		actual := DisplayNote(test.n)
		if actual != test.expected {
			t.Errorf("Expected %s, but was %s.", test.expected, actual)
		}
	}
}
