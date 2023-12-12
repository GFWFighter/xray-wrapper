package XRay

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"

	handlerService "github.com/xtls/xray-core/app/proxyman/command"
	"github.com/xtls/xray-core/common/errors"

	"github.com/xtls/xray-core/infra/conf"
	json_reader "github.com/xtls/xray-core/infra/conf/json"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func SwitchConfig(server string, config string) error {
	outbound := ParseOutbound(config)

	conn, err := grpc.Dial(server, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return err
	}
	defer conn.Close()

	client := handlerService.NewHandlerServiceClient(conn)

	rmReq := &handlerService.RemoveOutboundRequest{
		Tag: "proxy",
	}
	_, err = client.RemoveOutbound(context.Background(), rmReq)
	if err != nil {
		return err
	}

	confReader := strings.NewReader(outbound)
	jsonConfig, err := decodeJSONOutboundConfig(confReader)

	conf, err := jsonConfig.Build()
	if err != nil {
		return err
	}

	addReq := &handlerService.AddOutboundRequest{
		Outbound: conf,
	}

	_, err = client.AddOutbound(context.Background(), addReq)
	if err != nil {
		return err
	}

	return nil
}

type offset struct {
	line int
	char int
}

func findOffset(b []byte, o int) *offset {
	if o >= len(b) || o < 0 {
		return nil
	}

	line := 1
	char := 0
	for i, x := range b {
		if i == o {
			break
		}
		if x == '\n' {
			line++
			char = 0
		} else {
			char++
		}
	}

	return &offset{line: line, char: char}
}

type errPathObjHolder struct{}

func newError(values ...interface{}) *errors.Error {
	return errors.New(values...).WithPathObj(errPathObjHolder{})
}

func decodeJSONOutboundConfig(reader io.Reader) (*conf.OutboundDetourConfig, error) {
	jsonConfig := &conf.OutboundDetourConfig{}

	jsonContent := bytes.NewBuffer(make([]byte, 0, 10240))
	jsonReader := io.TeeReader(&json_reader.Reader{
		Reader: reader,
	}, jsonContent)
	decoder := json.NewDecoder(jsonReader)

	if err := decoder.Decode(jsonConfig); err != nil {
		var pos *offset
		cause := errors.Cause(err)
		switch tErr := cause.(type) {
		case *json.SyntaxError:
			pos = findOffset(jsonContent.Bytes(), int(tErr.Offset))
		case *json.UnmarshalTypeError:
			pos = findOffset(jsonContent.Bytes(), int(tErr.Offset))
		}
		if pos != nil {
			return nil, newError("failed to read config file at line ", pos.line, " char ", pos.char).Base(err)
		}
		return nil, newError("failed to read config file").Base(err)
	}

	return jsonConfig, nil
}
