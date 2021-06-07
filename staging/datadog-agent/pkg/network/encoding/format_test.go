package encoding

import (
	"testing"

	model "github.com/n9e/agent-payload/process"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/network"
	"github.com/stretchr/testify/require"
)

func TestFormatRouteIdx(t *testing.T) {

	tests := []struct {
		desc                string
		via                 *network.Via
		routesIn, routesOut map[string]RouteIdx
		idx                 int32
	}{
		{
			desc: "nil via",
			via:  nil,
			idx:  -1,
		},
		{
			desc: "empty via",
			via:  &network.Via{},
			idx:  -1,
		},
		{
			desc:     "empty routes, non-nil via",
			via:      &network.Via{Subnet: network.Subnet{Alias: "foo"}},
			idx:      0,
			routesIn: map[string]RouteIdx{},
			routesOut: map[string]RouteIdx{
				"foo": {Idx: 0, Route: model.Route{Subnet: &model.Subnet{Alias: "foo"}}},
			},
		},
		{
			desc: "non-empty routes, non-nil via with existing alias",
			via:  &network.Via{Subnet: network.Subnet{Alias: "foo"}},
			idx:  0,
			routesIn: map[string]RouteIdx{
				"foo": {Idx: 0, Route: model.Route{Subnet: &model.Subnet{Alias: "foo"}}},
			},
			routesOut: map[string]RouteIdx{
				"foo": {Idx: 0, Route: model.Route{Subnet: &model.Subnet{Alias: "foo"}}},
			},
		},
		{
			desc: "non-empty routes, non-nil via with non-existing alias",
			via:  &network.Via{Subnet: network.Subnet{Alias: "bar"}},
			idx:  1,
			routesIn: map[string]RouteIdx{
				"foo": {Idx: 0, Route: model.Route{Subnet: &model.Subnet{Alias: "foo"}}},
			},
			routesOut: map[string]RouteIdx{
				"foo": {Idx: 0, Route: model.Route{Subnet: &model.Subnet{Alias: "foo"}}},
				"bar": {Idx: 1, Route: model.Route{Subnet: &model.Subnet{Alias: "bar"}}},
			},
		},
	}

	for _, te := range tests {
		t.Run(te.desc, func(t *testing.T) {
			idx := formatRouteIdx(te.via, te.routesIn)
			require.Equal(t, te.idx, idx)
			require.Len(t, te.routesIn, len(te.routesIn), "routes in and out are not equal, expected: %v, actual: %v", te.routesOut, te.routesIn)
			for k, v := range te.routesOut {
				otherv, ok := te.routesIn[k]
				require.True(t, ok, "routes in and out are not equal, expected: %v, actual: %v", te.routesOut, te.routesIn)
				require.Equal(t, v, otherv, "routes in and out are not equal, expected: %v, actual: %v", te.routesOut, te.routesIn)
			}
		})
	}
}
