package XRay

import (
	"github.com/moqsien/vpnparser/pkgs/outbound"
	_ "github.com/moqsien/vpnparser/pkgs/outbound/xray"
	_ "github.com/moqsien/vpnparser/pkgs/parser"
)

func ParseOutbound(uri string) string {
	ob := outbound.GetOutbound(outbound.XrayCore, uri)
	if ob == nil {
		return ""
	}
	ob.Parse(uri)
	obStr := ob.GetOutboundStr()
	return obStr
}
