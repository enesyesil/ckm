package common

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type RawWorkload struct {
    ID       string `yaml:"id"`
    Type     string `yaml:"type"`
    CPUTime  string `yaml:"cpu_time"`
    MemoryMB int    `yaml:"memory_mb"`
	FilePath  string `yaml:"file_path"`
}

func LoadWorkloads(path string) ([]RawWorkload, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var raw []RawWorkload
    err = yaml.Unmarshal(data, &raw)
    if err != nil {
        return nil, err
    }

    return raw, nil
}

func ParseCPUTime(rawTime string) time.Duration {
    d, _ := time.ParseDuration(rawTime)
    return d
}
