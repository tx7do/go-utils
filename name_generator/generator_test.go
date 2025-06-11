package name_generator

import (
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	g := New()

	dictTypes := Scheme5

	result := g.Generate(dictTypes)

	if result == "" {
		t.Errorf("result is empty, please check the dictionary data")
	} else {
		t.Logf("generate`s nickname: %s", result)
	}
}

func TestGenerateParts(t *testing.T) {
	g := New()

	dictTypes := Scheme6

	parts := g.GenerateParts(dictTypes)

	if len(parts) == 0 {
		t.Errorf("result is empty, please check the dictionary data")
	} else {
		t.Logf("generate`s parts: %v", parts)
	}

	parts[0] = parts[0] + "#的"

	result := strings.Join(parts, "")
	t.Logf("generate`s nickname: %s", result)
}

func TestGenerateChineseName(t *testing.T) {
	g := New()

	result := g.GenerateChineseName(1, true, false)
	if result == "" {
		t.Errorf("result is empty, please check the dictionary data")
	} else {
		t.Logf("Generated single surname single name (female): %s", result)
	}

	result = g.GenerateChineseName(2, false, true)
	if result == "" {
		t.Errorf("result is empty, please check the dictionary data")
	} else {
		t.Logf("Generated compound surname double name (male): %s", result)
	}
}

func TestGenerateEnglishName(t *testing.T) {
	g := New()

	result := g.GenerateEnglishName(1, 0, 1, true)
	if result == "" {
		t.Errorf("result is empty, please check the dictionary data")
	} else {
		t.Logf("Generated female English name: %s", result)
	}

	result = g.GenerateEnglishName(2, 0, 1, false)
	if result == "" {
		t.Errorf("result is empty, please check the dictionary data")
	} else {
		t.Logf("Generated male English name: %s", result)
	}
}

func TestGenerateJapaneseNameCN(t *testing.T) {
	g := New()

	result := g.GenerateJapaneseNameCN()
	if result == "" {
		t.Errorf("result is empty, please check the dictionary data")
	} else {
		t.Logf("Generated Japanese name (CN): %s", result)
	}
}

func TestGenerateJapaneseName(t *testing.T) {
	g := New()

	result := g.GenerateJapaneseName()
	if result == "" {
		t.Errorf("result is empty, please check the dictionary data")
	} else {
		t.Logf("Generated Japanese name: %s", result)
	}
}
