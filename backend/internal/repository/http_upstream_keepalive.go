package repository

import "time"

// These are socket parameters, not a measured peer-disconnect deadline.
const (
	upstreamKeepAliveIdle     = 15 * time.Second
	upstreamKeepAliveInterval = 5 * time.Second
	upstreamKeepAliveCount    = 3
	upstreamTCPUserTimeout    = upstreamKeepAliveIdle + upstreamKeepAliveInterval*upstreamKeepAliveCount
)
