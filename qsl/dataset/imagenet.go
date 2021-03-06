package dataset

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	common "github.com/c3sr/dlframework/framework/predictor"
	"github.com/c3sr/dlframework/steps"
	"github.com/c3sr/pipeline"
)

type ImageNet struct {
	labels            []int
	names             []string
	dataPath          string
	dataInMemory      map[int]interface{}
	preprocessOptions common.PreprocessOptions
	preprocessMethod  string
}

func NewImageNet(dataPath string, dataList string, count int, preprocessOptions common.PreprocessOptions, preprocessMethod string) (*ImageNet, error) {

	start := time.Now()

	res := &ImageNet{
		dataPath:          dataPath,
		preprocessOptions: preprocessOptions,
		preprocessMethod:  preprocessMethod,
	}

	if dataList == "" {
		dataList = filepath.Join(dataPath, "val_map.txt")
	}
	file, err := os.Open(dataList)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	notFound := 0

	for scanner.Scan() {
		val := strings.Split(scanner.Text(), " ")
		if len(val) != 2 {
			return nil, fmt.Errorf("The format of the image list is not correct.")
		}

		imageName, label := val[0], val[1]

		src := filepath.Join(dataPath, imageName)
		if _, err := os.Stat(src); err != nil {
			notFound++
			continue
		}

		labelInt, err := strconv.Atoi(label)
		if err != nil {
			return nil, fmt.Errorf("Can't convert %s into an integer.", label)
		}

		res.names = append(res.names, imageName)
		res.labels = append(res.labels, labelInt)

		if count != 0 && len(res.labels) == count {
			break
		}
	}

	elapsed := time.Now().Sub(start)

	if len(res.labels) == 0 {
		return nil, fmt.Errorf("no images in image list found.")
	}
	if notFound > 0 {
		fmt.Printf("reduced image list, %d images not found.\n", notFound)
	}

	fmt.Printf("found %d images, took %.1f seconds.\n", len(res.labels), elapsed.Seconds())

	return res, nil

}

func (i *ImageNet) LoadQuerySamples(sampleList []int) error {
	i.dataInMemory = make(map[int]interface{})

	input := make(chan interface{}, defaultChannelBuffer)
	opts := []pipeline.Option{pipeline.ChannelBuffer(defaultChannelBuffer)}
	output := pipeline.New(opts...).
		Then(steps.NewPreprocessGeneral(i.preprocessOptions, i.preprocessMethod)).
		Run(input)

	for _, sample := range sampleList {
		input <- i.getItemLocation(sample)
	}

	close(input)

	for _, sample := range sampleList {
		i.dataInMemory[sample] = <-output
	}
	return nil
}

func (i *ImageNet) UnloadQuerySamples(sampleList []int) error {
	if i.dataInMemory == nil {
		return fmt.Errorf("Data map is nil.")
	}
	if len(sampleList) == 0 {
		i.dataInMemory = nil
	} else {
		for _, sample := range sampleList {
			delete(i.dataInMemory, sample)
		}
	}
	return nil
}

func (i *ImageNet) GetItemCount() int {
	return len(i.labels)
}

func (i *ImageNet) GetSamples(sampleList []int) (map[int]interface{}, error) {
	for _, sample := range sampleList {
		if _, exist := i.dataInMemory[sample]; !exist {
			return nil, fmt.Errorf("sample id %d not loaded.", sample)
		}
	}
	return i.dataInMemory, nil
}

func (i *ImageNet) getItemLocation(sample int) string {
	return filepath.Join(i.dataPath, i.names[sample])
}
