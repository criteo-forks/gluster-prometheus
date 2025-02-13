package glusterutils

import (
	"encoding/xml"

	"github.com/gluster/gluster-prometheus/pkg/glusterutils/glusterconsts"

	"fmt"
	"os/exec"
)

type healBricks struct {
	XMLName     xml.Name         `xml:"cliOutput"`
	Healentries []healEntriesXML `xml:"healInfo>bricks>brick"`
}

type healEntriesXML struct {
	XMLName        xml.Name `xml:"brick"`
	HostUUID       string   `xml:"hostUuid,attr"`
	Brickname      string   `xml:"name"`
	Connected      string   `xml:"status"`
	NumHealEntries string   `xml:"numberOfEntries"`
}

type gd1Brick struct {
	Name      string `xml:"name"`
	PeerID    string `xml:"hostUuid"`
	IsArbiter int    `xml:"IsArbiter"`
}

type gd1Option struct {
	Name  string `xml:"name"`
	Value string `xml:"value"`
}

type gd1Transport string

type gd1Volume struct {
	Name                    string       `xml:"name"`
	ID                      string       `xml:"id"`
	Status                  string       `xml:"statusStr"`
	Type                    string       `xml:"typeStr"`
	Bricks                  []gd1Brick   `xml:"bricks>brick"`
	NumBricks               int          `xml:"brickCount"`
	DistCount               int          `xml:"distCount"`
	ReplicaCount            int          `xml:"replicaCount"`
	DisperseCount           int          `xml:"disperseCount"`
	DisperseRedundancyCount int          `xml:"redundancyCount"`
	StripeCount             int          `xml:"stripeCount"`
	TransportRaw            gd1Transport `xml:"transport"`
	Options                 []gd1Option  `xml:"options>option"`
}

type gd1VolumesList struct {
	XMLName xml.Name `xml:"cliOutput"`
	List    []string `xml:"volList>volume"`
}

type gd1Volumes struct {
	XMLName xml.Name    `xml:"cliOutput"`
	List    []gd1Volume `xml:"volInfo>volumes>volume"`
}

type gd1VolumeStatusDetail struct {
	Name      string       `xml:"volName"`
	NodeCount int          `xml:"nodeCount"`
	Nodes     []gd1Process `xml:"node"`
}

type gd1VolumesDetail struct {
	XMLName xml.Name                `xml:"cliOutput"`
	List    []gd1VolumeStatusDetail `xml:"volStatus>volumes>volume"`
}

type gd1VolumeQuotaLimit struct {
	Path              string `xml:"path"`
	HardLimit         uint64 `xml:"hard_limit"`
	SoftLimitPercent  string `xml:"soft_limit_percent"`
	SoftLimitValue    uint64 `xml:"soft_limit_value"`
	UsedSpace         uint64 `xml:"used_space"`
	AvailSpace        uint64 `xml:"avail_space"`
	SoftLimitExceeded string `xml:"sl_exceeded"`
	HardLimitExceeded string `xml:"hl_exceeded"`
}

type gd1VolumeQuotas struct {
	XMLName xml.Name              `xml:"cliOutput"`
	List    []gd1VolumeQuotaLimit `xml:"volQuota>limit"`
}

type snapshotParentVolume struct {
	Name          string `xml:"name"`
	SnapCount     int    `xml:"snapCount"`
	SnapRemaining int    `xml:"snapRemaining"`
}

type snapshotVolume struct {
	Name         string               `xml:"name"`
	Status       string               `xml:"status"`
	OriginVolume snapshotParentVolume `xml:"originVolume"`
}

type gd1Snapshot struct {
	Name        string         `xml:"name"`
	UUID        string         `xml:"uuid"`
	Description string         `xml:"description"`
	CreateTime  string         `xml:"createTime"`
	VolCount    string         `xml:"volCount"`
	SnapVolume  snapshotVolume `xml:"snapVolume"`
}

type gd1Snapshots struct {
	XMLName xml.Name      `xml:"cliOutput"`
	List    []gd1Snapshot `xml:"snapInfo>snapshots>snapshot"`
}

type blockStat struct {
	Size   uint64 `xml:"size"`
	Reads  uint64 `xml:"reads"`
	Writes uint64 `xml:"writes"`
}

type gd1FopStat struct {
	Name       string  `xml:"name"`
	Hits       int     `xml:"hits"`
	AvgLatency float64 `xml:"avgLatency"`
	MinLatency float64 `xml:"minLatency"`
	MaxLatency float64 `xml:"maxLatency"`
}

type cumulativeStats struct {
	BlkStats   []blockStat  `xml:"blokcStats>block"`
	Duration   uint64       `xml:"duration"`
	TotalRead  uint64       `xml:"totalRead"`
	TotalWrite uint64       `xml:"totalWrite"`
	FopStats   []gd1FopStat `xml:"fopStats>fop"`
}

type intervalStats struct {
	BlkStats   []blockStat  `xml:"blokcStats>block"`
	Duration   uint64       `xml:"duration"`
	TotalRead  uint64       `xml:"totalRead"`
	TotalWrite uint64       `xml:"totalWrite"`
	FopStats   []gd1FopStat `xml:"fopStats>fop"`
}

type brickProfileInfo struct {
	Name     string          `xml:"brickName"`
	Stats    cumulativeStats `xml:"cumulativeStats"`
	IntStats intervalStats   `xml:"intervalStats"`
}

type volumeProfile struct {
	VolName    string             `xml:"volname"`
	ProfileOp  int                `xml:"profileOp"`
	BrickCount int                `xml:"brickCount"`
	Bricks     []brickProfileInfo `xml:"brick"`
}

type gd1ProfileInfo struct {
	XMLName    xml.Name      `xml:"cliOutput"`
	VolProfile volumeProfile `xml:"volProfile"`
}

type gd1ProtocolPorts struct {
	TCPPort  string `xml:"tcp"`
	RDMAPort string `xml:"rdma"`
}

type gd1Process struct {
	Hostname      string           `xml:"hostname"`
	Path          string           `xml:"path"`
	PeerID        string           `xml:"peerid"`
	Status        int              `xml:"status"`
	Port          string           `xml:"port"` // can contain 'N/A' entries
	ProtocolPorts gd1ProtocolPorts `xml:"ports"`
	PID           int              `xml:"pid"`
	InodesTotal   uint64           `xml:"inodesTotal"`
	InodesFree    uint64           `xml:"inodesFree"`
	SizeTotal     uint64           `xml:"sizeTotal"`
	SizeFree      uint64           `xml:"sizeFree"`
}

type gd1VolumeStatusInfo struct {
	Name          string       `xml:"volName"`
	NodeCount     int          `xml:"nodeCount"`
	NodeProcesses []gd1Process `xml:"node"`
}

type gd1VolumeStatus struct {
	XMLName xml.Name              `xml:"cliOutput"`
	List    []gd1VolumeStatusInfo `xml:"volStatus>volumes>volume"`
}

func (t *gd1Transport) String() string {
	// 0 - tcp
	// 1 - rdma
	// 2 - tcp,rdma
	if *t == "0" {
		return "tcp"
	} else if *t == "1" {
		return "rdma"
	}
	return "tcp,rdma"
}

func getSubvolType(voltype string) string {
	switch voltype {
	case glusterconsts.VolumeTypeDistReplicate:
		return glusterconsts.SubvolTypeReplicate
	case glusterconsts.VolumeTypeDistDisperse:
		return glusterconsts.SubvolTypeDisperse
	default:
		return voltype
	}
}

func getSubvolBricksCount(replicaCount int, disperseCount int) int {
	if replicaCount > 0 {
		return replicaCount
	}

	if disperseCount > 0 {
		return disperseCount
	}
	return 1
}

// execGluster runs `gluster` with --xml --remote-host=<...> and the args provided
func (g *GD1) execGluster(args ...string) ([]byte, error) {
	// always request output in XML format
	args = append(args, "--xml")
	// grab remote host from config
	if g.config.GlusterGlusterdSock != "" {
		args = append(args, fmt.Sprintf("--glusterd-sock=%s", g.config.GlusterGlusterdSock))
	} else if g.config.GlusterRemoteHost != "" {
		args = append(args, fmt.Sprintf("--remote-host=%s", g.config.GlusterRemoteHost))
	}
	return exec.Command(g.config.GlusterCmd, args...).Output() // #nosec
}
