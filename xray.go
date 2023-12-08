package XRay

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	runtimeDebug "runtime/debug"

	L "github.com/gfwfighter/xray-wrapper/log"
	"github.com/xtls/xray-core/common/log"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf/serial"

	_ "github.com/xtls/xray-core/main/distro/all"

	_ "golang.org/x/mobile/bind"
)

type Instance struct {
	instance   *core.Instance
	configPath string
	logger     *L.Logger
}

type Logger interface {
	L.RawLogger
}

func registerLoader() {
	core.RegisterConfigLoader(&core.ConfigFormat{
		Name:      "JSON",
		Extension: []string{"json"},
		Loader: func(input interface{}) (*core.Config, error) {
			switch v := input.(type) {
			case io.Reader:
				return serial.LoadJSONConfig(v)
			default:
				panic("unsupported")
			}
		},
	})
}

func NewInstance(configPath string, assetPath string, logger Logger) (*Instance, error) {
	os.Setenv("XRAY_LOCATION_ASSET", assetPath)
	os.Setenv("XRAY_LOCATION_CONFIG", filepath.Dir(configPath))
	file, err := os.OpenFile(configPath, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	registerLoader()
	config, err := core.LoadConfig("json", file)
	if err != nil {
		return nil, err
	}
	instance, err := core.New(config)
	if err != nil {
		return nil, err
	}
	logContainer := L.New(logger)
	log.RegisterHandler(logContainer)
	return &Instance{
		instance:   instance,
		configPath: configPath,
		logger:     logContainer,
	}, nil
}

func (i *Instance) Start() error {
	err := i.instance.Start()
	if err != nil {
		return err
	}
	runtime.GC()
	runtimeDebug.FreeOSMemory()
	return nil
}

func (i *Instance) Stop() error {
	err := i.instance.Close()
	if err != nil {
		return err
	}
	return nil
}
