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

	_ "github.com/xtls/xray-core/app/dispatcher"
	_ "github.com/xtls/xray-core/app/dns"
	_ "github.com/xtls/xray-core/app/dns/fakedns"
	_ "github.com/xtls/xray-core/app/log"
	_ "github.com/xtls/xray-core/app/metrics"
	_ "github.com/xtls/xray-core/app/observatory"
	_ "github.com/xtls/xray-core/app/policy"
	_ "github.com/xtls/xray-core/app/proxyman/inbound"
	_ "github.com/xtls/xray-core/app/proxyman/outbound"
	_ "github.com/xtls/xray-core/app/router"
	_ "github.com/xtls/xray-core/app/stats"
	_ "github.com/xtls/xray-core/proxy/blackhole"
	_ "github.com/xtls/xray-core/proxy/dns"
	_ "github.com/xtls/xray-core/proxy/dokodemo"
	_ "github.com/xtls/xray-core/proxy/freedom"
	_ "github.com/xtls/xray-core/proxy/http"
	_ "github.com/xtls/xray-core/proxy/shadowsocks"
	_ "github.com/xtls/xray-core/proxy/socks"
	_ "github.com/xtls/xray-core/proxy/trojan"
	_ "github.com/xtls/xray-core/proxy/vless/outbound"
	_ "github.com/xtls/xray-core/proxy/vmess/outbound"
	_ "github.com/xtls/xray-core/proxy/wireguard"
	_ "github.com/xtls/xray-core/transport/internet/grpc"
	_ "github.com/xtls/xray-core/transport/internet/headers/http"
	_ "github.com/xtls/xray-core/transport/internet/headers/noop"
	_ "github.com/xtls/xray-core/transport/internet/headers/srtp"
	_ "github.com/xtls/xray-core/transport/internet/headers/tls"
	_ "github.com/xtls/xray-core/transport/internet/headers/utp"
	_ "github.com/xtls/xray-core/transport/internet/headers/wechat"
	_ "github.com/xtls/xray-core/transport/internet/headers/wireguard"
	_ "github.com/xtls/xray-core/transport/internet/http"
	_ "github.com/xtls/xray-core/transport/internet/kcp"
	_ "github.com/xtls/xray-core/transport/internet/quic"
	_ "github.com/xtls/xray-core/transport/internet/reality"
	_ "github.com/xtls/xray-core/transport/internet/tagged/taggedimpl"
	_ "github.com/xtls/xray-core/transport/internet/tcp"
	_ "github.com/xtls/xray-core/transport/internet/tls"
	_ "github.com/xtls/xray-core/transport/internet/udp"
	_ "github.com/xtls/xray-core/transport/internet/websocket"

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
