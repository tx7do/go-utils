package bank_card

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNameOfBank(t *testing.T) {
	var validBankCards = []string{
		"6226095711989751",
		"6228480402564890018",
		"6228480402637874213",
		"6228481552887309119",
		"6228480801416266113",
		"6228481698729890079",
		"621661280000447287",
		"6222081106004039591",
	}

	for _, w := range validBankCards {
		t.Run("get bank card of name: "+w, func(t *testing.T) {
			name := GetNameOfBank(w)
			fmt.Println(w, name)
			assert.True(t, len(name) > 0)
		})
	}
}
