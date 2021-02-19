package mlcommonsmlmomodelscope

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/c3sr/config"
	_ "github.com/c3sr/tracer/jaeger"
)

func TestInfoModels(t *testing.T) {
	// if err := InfoModels("onnxruntime"); err != nil {
	// 	t.Error(err)
	// }

	// if err := InfoModels("mxnet"); err != nil {
	// 	t.Error(err)
	// }

	// if err := InfoModels("tensorflow"); err != nil {
	// 	t.Error(err)
	// }

	// if err := InfoModels("unimplemented"); err == nil {
	// 	t.Errorf("backend not implemented but found.")
	// }
}

func TestInterfaces(t *testing.T) {
	pwd, _ := os.Getwd()
	os.Setenv("DATA_DIR", filepath.Join(pwd, "qsl/dataset/_fixtures/fake_imagenet"))
	if err := Initialize("pytorch", "torchvision_alexnet", "1.0",
		"imagenet", "", 10, false, "FULL_TRACE"); err != nil {
		t.Error(err)
	}

	if err := LoadQuerySamples([]int{0, 1, 2, 3, 4, 5, 6, 7}); err != nil {
		t.Error(err)
	}

	if err := UnloadQuerySamples([]int{1, 3, 5, 7}); err != nil {
		t.Error(err)
	}

	res, err := IssueQuery([]int{0, 2, 4, 6})
	if err != nil {
		t.Error(err)
	}

	for _, out := range res {
		fmt.Println("Index: ", out[0].GetClassification().GetIndex())
		fmt.Println("Label: ", out[0].GetClassification().GetLabel())
		fmt.Println("Probability: ", out[0].GetProbability())
	}

	if err := UnloadQuerySamples([]int{}); err != nil {
		t.Error(err)
	}

	if err := Finalize(); err != nil {
		t.Error(err)
	}
}

func TestMain(m *testing.M) {
	config.Init(
		config.AppName("mlcommons-mlmodelscope"),
		config.DebugMode(true),
		config.VerboseMode(true),
	)
	os.Exit(m.Run())
}