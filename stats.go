package xray

import (
	"context"
	"fmt"
	"path"
	"reflect"

	statsService "github.com/xtls/xray-core/app/stats/command"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// query system stats and outbound stats.
// server means The API server address, like "127.0.0.1:8080".
// dir means the dir which result json will be wrote to.
func QueryStats(server string) string {
	conn, err := grpc.Dial(server, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return err.Error()
	}
	defer conn.Close()

	client := statsService.NewStatsServiceClient(conn)

	sysStatsReq := &statsService.SysStatsRequest{}
	sysStatsRes, err := client.GetSysStats(context.Background(), sysStatsReq)
	if err != nil {
		return err.Error()
	}
	sysStatsPath := path.Join(dir, "sysStats.json")
	err = writeResult(sysStatsRes, sysStatsPath)
	if err != nil {
		return err.Error()
	}

	statsReq := &statsService.QueryStatsRequest{
		Pattern: "",
		Reset_:  false,
	}
	statsRes, err := client.QueryStats(context.Background(), statsReq)
	if err != nil {
		return err.Error()
	}
	statsPath := path.Join(dir, "stats.json")
	return statsRes
}
