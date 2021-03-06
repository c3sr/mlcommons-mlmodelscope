package dataset

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	common "github.com/c3sr/dlframework/framework/predictor"
	"github.com/c3sr/dlframework/steps"
	"github.com/c3sr/pipeline"
)

type Coco struct {
	labels            []int
	names             []string
	bboxes            [][4]float64
	dataPath          string
	dataInMemory      map[int]interface{}
	preprocessOptions common.PreprocessOptions
	preprocessMethod  string
}

func NewCoco(dataPath string, dataList string, count int, preprocessOptions common.PreprocessOptions, preprocessMethod string) (*Coco, error) {

	start := time.Now()

	res := &Coco{
		dataPath:          dataPath,
		preprocessOptions: preprocessOptions,
		preprocessMethod:  preprocessMethod,
	}

	if dataList == "" {
		dataList = filepath.Join(dataPath, "annotations/instances_val2017.json")
	}
	file, err := os.Open(dataList)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var f interface{}

	jsonBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(jsonBytes, &f)

	coco := f.(map[string]interface{})

	mapToIdx := make(map[int]int)

	descriptions := coco["images"].([]interface{})

	annotations := coco["annotations"].([]interface{})

	notFound := 0

	for _, rawdesc := range descriptions {
		desc := rawdesc.(map[string]interface{})

		imageName := filepath.Join("val2017", desc["file_name"].(string))

		src := filepath.Join(dataPath, imageName)
		if _, err := os.Stat(src); err != nil {
			notFound++
			continue
		}

		mapToIdx[int(desc["id"].(float64))] = len(res.names)
		res.names = append(res.names, imageName)

		if count != 0 && len(res.names) == count {
			break
		}
	}

	res.labels = make([]int, len(res.names))
	res.bboxes = make([][4]float64, len(res.names))

	for _, rawAnnotation := range annotations {
		anno := rawAnnotation.(map[string]interface{})
		i, ok := mapToIdx[int(anno["image_id"].(float64))]
		if !ok {
			continue
		}
		res.labels[i] = int(anno["category_id"].(float64))
		bbox := anno["bbox"].([]interface{})
		for j := 0; j < 4; j++ {
			res.bboxes[i][j] = bbox[j].(float64)
		}
	}

	elapsed := time.Now().Sub(start)

	if len(res.names) == 0 {
		return nil, fmt.Errorf("no images in image list found.")
	}
	if notFound > 0 {
		fmt.Printf("reduced image list, %d images not found.\n", notFound)
	}

	fmt.Printf("found %d images, took %.1f seconds.\n", len(res.names), elapsed.Seconds())

	return res, nil
}

func (c *Coco) LoadQuerySamples(sampleList []int) error {
	c.dataInMemory = make(map[int]interface{})

	input := make(chan interface{}, defaultChannelBuffer)
	opts := []pipeline.Option{pipeline.ChannelBuffer(defaultChannelBuffer)}
	output := pipeline.New(opts...).
		Then(steps.NewPreprocessGeneral(c.preprocessOptions, c.preprocessMethod)).
		Run(input)

	for _, sample := range sampleList {
		input <- c.getItemLocation(sample)
	}

	close(input)

	for _, sample := range sampleList {
		c.dataInMemory[sample] = <-output
	}
	return nil
}

func (c *Coco) UnloadQuerySamples(sampleList []int) error {
	if c.dataInMemory == nil {
		return fmt.Errorf("Data map is nil.")
	}
	if len(sampleList) == 0 {
		c.dataInMemory = nil
	} else {
		for _, sample := range sampleList {
			delete(c.dataInMemory, sample)
		}
	}
	return nil
}

func (c *Coco) GetItemCount() int {
	return len(c.labels)
}

func (c *Coco) GetSamples(sampleList []int) (map[int]interface{}, error) {
	for _, sample := range sampleList {
		if _, exist := c.dataInMemory[sample]; !exist {
			return nil, fmt.Errorf("sample id %d not loaded.", sample)
		}
	}
	return c.dataInMemory, nil
}

func (c *Coco) getItemLocation(sample int) string {
	return filepath.Join(c.dataPath, c.names[sample])
}
