package XRay

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime"
	runtimeDebug "runtime/debug"
	"strconv"
	"strings"

	L "github.com/gfwfighter/xray-wrapper/log"
	"github.com/xtls/xray-core/common/log"
	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf"
	J "github.com/xtls/xray-core/infra/conf/json"
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
	port       int
	apiPort    int
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

func getPort() (int, int, error) {
	listener, err := net.Listen("tcp", "[::1]:0")
	if err != nil {
		return -1, -1, err
	}
	listener2, err := net.Listen("tcp", "[::1]:0")
	if err != nil {
		return -1, -1, err
	}
	err = listener.Close()
	if err != nil {
		return -1, -1, err
	}
	err = listener2.Close()
	if err != nil {
		return -1, -1, err
	}
	return listener.Addr().(*net.TCPAddr).Port, listener2.Addr().(*net.TCPAddr).Port, nil
}

func decodeJSON(config string) (*core.InboundHandlerConfig, error) {
	reader := strings.NewReader(config)
	jsonConfig := &conf.InboundDetourConfig{}
	jsonContent := bytes.NewBuffer(make([]byte, 0, 10240))
	jsonReader := io.TeeReader(&J.Reader{
		Reader: reader,
	}, jsonContent)
	decoder := json.NewDecoder(jsonReader)
	err := decoder.Decode(jsonConfig)
	if err != nil {
		return nil, err
	}
	result, err := jsonConfig.Build()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func getInboundConfig(port int) string {
	return "{\"listen\":\"[::1]\",\"port\":" + strconv.Itoa(port) +
		",\"protocol\":\"socks\",\"tag\":\"socks-in\"," +
		"\"settings\":{\"auth\":\"noauth\",\"udp\":true}," +
		"\"sniffing\":{\"enabled\":true,\"destOverride\": [\"fakedns+others\"]}}"
}

func getAPInboundConfig(port int) string {
	return "{\"listen\":\"127.0.0.1\",\"port\":" + strconv.Itoa(port) +
		",\"protocol\":\"dokodemo-door\",\"tag\":\"api\"," +
		"\"settings\":{\"address\":\"127.0.0.1\"}}"
}

func NewInstance(configPath string, assetPath string, enableAPI bool, logger Logger) (*Instance, error) {
	os.Setenv("XRAY_LOCATION_ASSET", assetPath)
	os.Setenv("XRAY_LOCATION_CONFIG", configPath)
	port, apiPort, err := getPort()
	file, err := os.OpenFile(filepath.Join(configPath, "config.json"), os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	config, err := core.LoadConfig("json", file)
	if err != nil {
		return nil, err
	}
	json, err := decodeJSON(getInboundConfig(port))
	if err != nil {
		return nil, err
	}
	if enableAPI {
		apiJson, err := decodeJSON(getAPInboundConfig(apiPort))
		if err != nil {
			return nil, err
		}
		config.Inbound = []*core.InboundHandlerConfig{json, apiJson}
	} else {
		apiPort = -1
		config.Inbound = []*core.InboundHandlerConfig{json}
	}
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
		port:       port,
		logger:     logContainer,
		apiPort:    apiPort,
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

func (i *Instance) GetPort() int {
	return i.port
}
