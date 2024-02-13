package config

import (
	"math/rand"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindMinimalIPNet(t *testing.T) {
	testCases := []struct {
		ipRanges []string
		want     string
	}{
		{
			ipRanges: []string{"192.168.0.1/32", "192.168.0.2/32", "192.168.0.3/32", "192.168.0.4/32"},
			want:     "192.168.0.0/24",
		},
		{
			ipRanges: []string{"192.168.0.1/32", "192.168.0.2/24"},
			want:     "192.168.0.0/24",
		},
		{
			ipRanges: []string{"192.168.0.1/32", "192.168.0.2/8"},
			want:     "192.0.0.0/8",
		},
		{
			ipRanges: []string{"192.168.0.1/32", "1.1.1.1/32"},
			want:     "0.0.0.0/0",
		},
		{
			ipRanges: []string{"192.168.0.1/31", "192.168.0.128/32"},
			want:     "192.168.0.0/24",
		},
		{
			ipRanges: []string{"192.168.0.1/32", "192.168.1.1/32"},
			want:     "192.168.0.0/16",
		},
	}
	for _, tc := range testCases {
		foundNet, err := findMinimalIPNet(tc.ipRanges)
		require.NoError(t, err)
		_, wantedNet, err := net.ParseCIDR(tc.want)
		require.NoError(t, err)
		require.Equal(t, wantedNet, foundNet)
	}
}

func TestGenerateRandomIP(t *testing.T) {
	rnd := rand.New(rand.NewSource(1337))
	ipRanges := []string{"192.168.0.1/32", "192.168.0.2/32"}
	minNet, _ := findMinimalIPNet(ipRanges)
	randIP, minNetStr, err := generateRandomIP(rnd, minNet, ipRanges)
	require.NoError(t, err)
	require.Equal(t, minNet.String(), minNetStr)
	require.Equal(t, "192.168.0.0/24", minNetStr)
	require.Equal(t, "192.168.0.42/32", randIP)
}

func TestGenerateRandomIPLoop(t *testing.T) {
	rnd := rand.New(rand.NewSource(1337))
	ipRanges := []string{"192.168.0.1/32", "192.168.0.2/32"}
	minNet, _ := findMinimalIPNet(ipRanges)
	for {
		randIP, _, genErr := generateRandomIP(rnd, minNet, ipRanges)
		require.NoError(t, genErr)
		require.NotEmptyf(t, randIP, "randIP is empty")
		ipRanges = append(ipRanges, randIP)
		if len(ipRanges) == 254 {
			break
		}
	}

	// we have fully used the subnet, next random IP should be empty
	randIP, _, err := generateRandomIP(rnd, minNet, ipRanges)
	require.NoError(t, err)
	require.Equal(t, "", randIP)
}
