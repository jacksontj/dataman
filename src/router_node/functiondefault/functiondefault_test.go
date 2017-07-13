package functiondefault

import (
	"context"
	"testing"
	"time"

	"github.com/jacksontj/dataman/src/datamantype"
)

type functionDefaultTestCase struct {
	fdName       string
	globalArgs   map[string]interface{}
	instanceArgs map[string]interface{}
}

var functionDefaultTestCases []*functionDefaultTestCase

func init() {
	functionDefaultTestCases = []*functionDefaultTestCase{
		// TODO: just iterate over all of them
		&functionDefaultTestCase{
			fdName: "uuid4",
		},
		&functionDefaultTestCase{
			fdName: "random",
		},
		&functionDefaultTestCase{
			fdName: "ksuid",
		},
	}
}

func testFunctionDefault(t *testing.T, fd FunctionDefault, datamanType datamantype.DatamanType) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	for i := 0; i < 100; i++ {
		val, err := fd.GetDefault(ctx, datamanType)
		if err != nil {
			t.Fatalf("Error getting value: %v", err)
		}
		_, err = datamanType.Normalize(val)
		if err != nil {
			t.Fatalf("Error normalizing value: %v", err)
		}
	}
}

func TestFunctionDefault(t *testing.T) {
	for _, testCase := range functionDefaultTestCases {
		fd := FunctionDefaultType(testCase.fdName).Get()

		if err := fd.Init(testCase.globalArgs, testCase.instanceArgs); err != nil {
			t.Fatalf("Error: %v", err)
		}
		// For each case, run something
		t.Run(testCase.fdName, func(t *testing.T) {
			for _, datamanType := range fd.SupportedTypes() {

				// For each type, run it again
				t.Run(string(datamanType), func(t *testing.T) {
					testFunctionDefault(t, fd, datamanType)
				})
			}
		})
	}

}
