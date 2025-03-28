package calculator

import (
	"context"
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	ctx := context.Background()
	tool := New()
	ret := new(Output)
	if err := tool.Run(ctx, NewInput("2+2", nil), ret); err != nil {
		t.Error(err)
	}
	switch value := ret.Result.(type) {
	case float64:
		if int(value) != 4 {
			t.Errorf("expecting 4, but got %.2f", value)
		}
	case int, int32, int64:
		t.Error("expecting float64, but got int")
	case bool:
		t.Error("expecting float64, but got bool")
	case string:
		t.Error("expecting float64, but got string")
	}
}

func ExampleCalculator() {
	ctx := context.Background()
	tool := New()
	ret := new(Output)
	tool.Run(ctx, NewInput("2+2", nil), ret)
	fmt.Println(ret.Result)
	// Output:
	// 4
}
