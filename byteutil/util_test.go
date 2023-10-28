package byteutil

import (
	"fmt"
	"testing"
)

func TestIntToBytes(t *testing.T) {
	fmt.Println(IntToBytes(1))
	fmt.Println(BytesToInt(IntToBytes(1)))
}
