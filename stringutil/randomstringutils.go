package stringutil

import (
	"fmt"
	"math"
	"math/rand"
	"time"
	"unicode"
)

var RANDOM = rand.New(rand.NewSource(time.Now().UnixNano()))

func RandomNonAlphaNumeric(count int) (string, error) {
	return RandomAlphaNumericCustom(count, false, false)
}

func RandomAscii(count int) (string, error) {
	return Random(count, 32, 127, false, false)
}

func RandomNumeric(count int) (string, error) {
	return Random(count, 0, 0, false, true)
}

func RandomAlphabetic(count int) (string, error) {
	return Random(count, 0, 0, true, false)
}

func RandomAlphaNumeric(count int) (string, error) {
	return Random(count, 0, 0, true, true)
}

func RandomAlphaNumericCustom(count int, letters bool, numbers bool) (string, error) {
	return Random(count, 0, 0, letters, numbers)
}

func Random(count int, start int, end int, letters bool, numbers bool, chars ...rune) (string, error) {
	return RandomSeed(count, start, end, letters, numbers, chars, RANDOM)
}

func RandomSeed(count int, start int, end int, letters bool, numbers bool, chars []rune, random *rand.Rand) (string, error) {

	if count == 0 {
		return "", nil
	} else if count < 0 {
		err := fmt.Errorf("randomstringutils illegal argument: Requested random string length %v is less than 0", count) // equiv to err := errors.New("...")
		return "", err
	}
	if chars != nil && len(chars) == 0 {
		err := fmt.Errorf("randomstringutils illegal argument: The chars array must not be empty")
		return "", err
	}

	if start == 0 && end == 0 {
		if chars != nil {
			end = len(chars)
		} else {
			if !letters && !numbers {
				end = math.MaxInt32
			} else {
				end = 'z' + 1
				start = ' '
			}
		}
	} else {
		if end <= start {
			err := fmt.Errorf("randomstringutils illegal argument: Parameter end (%v) must be greater than start (%v)", end, start)
			return "", err
		}

		if chars != nil && end > len(chars) {
			err := fmt.Errorf("randomstringutils illegal argument: Parameter end (%v) cannot be greater than len(chars) (%v)", end, len(chars))
			return "", err
		}
	}

	buffer := make([]rune, count)
	gap := end - start

	// high-surrogates range, (\uD800-\uDBFF) = 55296 - 56319
	//  low-surrogates range, (\uDC00-\uDFFF) = 56320 - 57343

	for count != 0 {
		count--
		var ch rune
		if chars == nil {
			ch = rune(random.Intn(gap) + start)
		} else {
			ch = chars[random.Intn(gap)+start]
		}

		if letters && unicode.IsLetter(ch) || numbers && unicode.IsDigit(ch) || !letters && !numbers {
			if ch >= 56320 && ch <= 57343 { // low surrogate range
				if count == 0 {
					count++
				} else {
					// Insert low surrogate
					buffer[count] = ch
					count--
					// Insert high surrogate
					buffer[count] = rune(55296 + random.Intn(128))
				}
			} else if ch >= 55296 && ch <= 56191 { // High surrogates range (Partial)
				if count == 0 {
					count++
				} else {
					// Insert low surrogate
					buffer[count] = rune(56320 + random.Intn(128))
					count--
					// Insert high surrogate
					buffer[count] = ch
				}
			} else if ch >= 56192 && ch <= 56319 {
				// private high surrogate, skip it
				count++
			} else {
				// not one of the surrogates*
				buffer[count] = ch
			}
		} else {
			count++
		}
	}
	return string(buffer), nil
}
