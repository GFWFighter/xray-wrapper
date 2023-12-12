package XRay

import (
	"context"

	statsService "github.com/xtls/xray-core/app/stats/command"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ProxyStats struct {
	Uplink   int64
	Downlink int64
}

func GetProxyStats(server string) (*ProxyStats, error) {
	conn, err := grpc.Dial(server, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := statsService.NewStatsServiceClient(conn)

	upReq := &statsService.GetStatsRequest{
		Name:   "outbound>>>proxy>>>traffic>>>uplink",
		Reset_: false,
	}
	upRes, err := client.GetStats(context.Background(), upReq)
	if err != nil {
		return nil, err
	}

	dnReq := &statsService.GetStatsRequest{
		Name:   "outbound>>>proxy>>>traffic>>>downlink",
		Reset_: false,
	}
	dnRes, err := client.GetStats(context.Background(), dnReq)
	if err != nil {
		return nil, err
	}

	return &ProxyStats{
		Uplink:   upRes.GetStat().GetValue(),
		Downlink: dnRes.GetStat().GetValue(),
	}, nil
}
