package slug

import "testing"

func TestSlugGeneration(t *testing.T) {
	testCases := map[string]string{
		"one two":                "one-two",
		"three four":             "three-four",
		"i have stop words here": "stop-words",
		"un texte générique comme 'Du texte. Du texte": "un-texte-generique-comme-du-texte-du-texte",
		"რომ გვერდის წაკითხვად":                        "rom-gverdis-cakit-xvad",
		"यह एक लंबा":                                   "yh-ek-lnbaa",
		"यह एक लंबा टेक्स्ट है":                        "yh-ek-lnbaa-ttekstt-hai",
	}

	for input, expected := range testCases {
		actual, err := New(input)
		if err != nil {
			t.Errorf("New(%q) returned an error: %v", input, err)
		}
		if actual != expected {
			t.Errorf("Slug(%q) = %q, want %q", input, actual, expected)
		}
	}
}

func TestSkipWords(t *testing.T) {
	input := "omit words that are known"
	expected := "words-known"
	actual, err := New(input, "OMIT That")
	if err != nil {
		t.Errorf("New(%q) returned an error: %v", input, err)
	}
	if actual != expected {
		t.Errorf("Slug(%q) = %q, want %q", input, actual, expected)
	}
}
