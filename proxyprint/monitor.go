package main

import (
	"encoding/json"
	"sync"
	"sync/atomic"
)

type Monitor struct {
	// The current number of clients connected (to the "listen" addr).
	CurrentClients AtomicInt64 `json:"currentClients"`
	// The total number of clients that have ever connected.
	TotalClients AtomicUint64 `json:"totalClients"`
	// The total number of failed attempts to connect to the server (to the
	// "connect" addr).
	TotalConnectServerFails AtomicUint64 `json:"totalConnectServerFails"`

	// The total number of attempts to dial tunnels to remote.
	TotalTunnelConnectAttempts AtomicUint64 `json:"totalTunnelConnectAttempts"`
	// The total number of tunnels connected to remote (not having gone through
	// the process of tunneling).
	TotalTunnelsConnected AtomicUint64 `json:"totalTunnelAttempts"`
	TunnelsAtMaxErr       AtomicBool   `json:"tunnelsAtMaxErr"`
	// The current number of tunnels to remote (to the "tunnel" addr).
	CurrentTunnels AtomicInt64 `json:"currentTunnels"`
	// The total number of tunnels to remote (ever).
	TotalTunnels AtomicUint64 `json:"totalTunnels"`

	// The total number of times a timeout occurred while trying to pipe to a
	// tunnel from client.
	TotalTunnelWaitTimeouts AtomicUint64 `json:"totalTunnelWaitTimeouts"`

	// The total number of server connections accepted for tunneling (accepted
	// from Accept(), not that it was "accepted" as a valid tunnel).
	TotalAcceptedServers AtomicUint64 `json:"totalAcceptedServers"`
	// The current number of tunnels from servers (to the "listen-servers" addr).
	CurrentTunneled AtomicInt64 `json:"currentTunneled"`
	// The total number of tunnels from servers (ever).
	TotalTunneled AtomicUint64 `json:"totalTunneled"`
	// The total number of tunnels from servers that failed readiness check.
	TotalTunneledFailedReady AtomicUint64 `json:"totalTunneledFailedReady"`

	// The config that is being used
	Config Config `json:"config"`

	ShuttingDown AtomicBool `json:"shuttingDown"`
	wg           sync.WaitGroup
}

func (mtr *Monitor) AddClient() (int64, uint64) {
	mtr.wg.Add(1)
	c := mtr.CurrentClients.Add(1)
	t := mtr.TotalClients.Add(1)
	return c, t
}
func (mtr *Monitor) RemoveClient() int64 {
	mtr.wg.Done()
	return mtr.CurrentClients.Add(-1)
}

func (mtr *Monitor) AddTotalConnectServerFails(_ error) uint64 {
	return mtr.TotalConnectServerFails.Add(1)
}

func (mtr *Monitor) AddTotalTunnelConnectAttempts() uint64 {
	return mtr.TotalTunnelConnectAttempts.Add(1)
}

func (mtr *Monitor) AddTotalTunnelsConnected() uint64 {
	return mtr.TotalTunnelsConnected.Add(1)
}

func (mtr *Monitor) AddTunnel() (int64, uint64) {
	mtr.wg.Add(1)
	c := mtr.CurrentTunnels.Add(1)
	t := mtr.TotalTunnels.Add(1)
	return c, t
}
func (mtr *Monitor) RemoveTunnel() int64 {
	mtr.wg.Done()
	return mtr.CurrentTunnels.Add(-1)
}

func (mtr *Monitor) AddTunnelWaitTimeouts() uint64 {
	return mtr.TotalTunnelWaitTimeouts.Add(1)
}

func (mtr *Monitor) AddTotalAcceptedServers() uint64 {
	return mtr.TotalAcceptedServers.Add(1)
}

func (mtr *Monitor) AddTunneled() (int64, uint64) {
	mtr.wg.Add(1)
	c := mtr.CurrentTunneled.Add(1)
	t := mtr.TotalTunneled.Add(1)
	return c, t
}
func (mtr *Monitor) RemoveTunneled() int64 {
	mtr.wg.Done()
	return mtr.CurrentTunneled.Add(-1)
}

func (mtr *Monitor) AddTunneledFailedReady() uint64 {
	return mtr.TotalTunneledFailedReady.Add(1)
}

func (mtr *Monitor) Wait() {
	mtr.wg.Wait()
}

type AtomicInt64 struct {
	atomic.Int64
}

func (a *AtomicInt64) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Load())
}

type AtomicUint64 struct {
	atomic.Uint64
}

func (a *AtomicUint64) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Load())
}

type AtomicBool struct {
	atomic.Bool
}

func (a *AtomicBool) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Load())
}
