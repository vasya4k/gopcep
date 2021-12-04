package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"gopcep/pcep"
	"strings"
	"sync"

	"github.com/golang/protobuf/ptypes"
	api "github.com/osrg/gobgp/api"
	gobgp "github.com/osrg/gobgp/pkg/server"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/anypb"
)

func printAsJSON(i interface{}) {
	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		logrus.Println(err)
	}
	fmt.Println(string(b))
}

type Node struct {
	ASN          uint32
	RouterID     string
	IGPRouteID   string
	Name         string
	Age          string
	SRRangeStart int
	SRRangeEnd   int
}

type Link struct {
	LocalNode       string
	RemoteNode      string
	IntIP           string
	NeighbourIP     string
	DefaultTEMetric uint32
	IGPMetric       uint32
	BW              float32
	ReservableBW    float32
	UnreservedBW    float32
	SRAdjacencySID  uint32
}

type Prefix struct {
	Prefix      string
	SRPrefixSID uint32
	LocalNode   string
}

type TopoView struct {
	*sync.RWMutex
	NodesByRouterID    map[string]*Node
	NodesByIGPRouteID  map[string]*Node
	LinksByIGPRouteID  []*Link
	PrefixByIGPRouteID map[string]*Prefix
	Paths              map[string][]*Path
	TopologyUpdate     chan bool `json:"-"`
}

type PathID struct {
	Src string
	Dst string
}

type Path struct {
	Src   string
	Dst   string
	Cost  int
	Links []*Link
}

func NewTopoView() *TopoView {
	return &TopoView{
		NodesByRouterID:    make(map[string]*Node),
		NodesByIGPRouteID:  make(map[string]*Node),
		LinksByIGPRouteID:  make([]*Link, 0),
		PrefixByIGPRouteID: make(map[string]*Prefix),
		Paths:              make(map[string][]*Path),
		TopologyUpdate:     make(chan bool),
		RWMutex:            &sync.RWMutex{},
	}
}

func getLinksCopyWithRemovedElement(fromLinks []*Link, index int) []*Link {
	links := make([]*Link, len(fromLinks))
	copy(links, fromLinks)
	return append(links[:index], links[index+1:]...)
}

func (t *TopoView) FindPathsForAllSrcDstPairs() {
	defer t.Unlock()

	t.Lock()
	for srcNode := range t.NodesByIGPRouteID {
		for dstNode := range t.NodesByIGPRouteID {
			if srcNode != dstNode {
				t.FindAllPathsForSrcDst(srcNode, dstNode)
			}
		}
	}
}

func (t *TopoView) FindAllPathsForSrcDst(src, dst string) {
	for i, link := range t.LinksByIGPRouteID {
		path := Path{
			Src:   src,
			Dst:   dst,
			Links: make([]*Link, 0),
		}
		if link.LocalNode == src && link.RemoteNode == dst {
			path.Links = append(path.Links, link)
			path.Cost = path.Cost + int(link.IGPMetric)
			pID := src + ":" + dst
			paths, ok := t.Paths[pID]
			if ok {
				paths = append(paths, &path)
				t.Paths[pID] = paths
				continue
			}
			t.Paths[pID] = []*Path{
				0: &path,
			}
			continue
		}
		// Found starting point
		if link.LocalNode == src {
			// path = append(path, link)
			t.FindPath(src, dst, link, &path, getLinksCopyWithRemovedElement(t.LinksByIGPRouteID, i))
		}
	}

}

func (t *TopoView) FindPath(src, dst string, previousLink *Link, path *Path, links []*Link) {

	for i, link := range links {
		if previousLink.LocalNode == link.RemoteNode {
			continue
		}

		if previousLink.RemoteNode == link.LocalNode && link.RemoteNode == dst {
			path.Links = append(path.Links, previousLink, link)

			path.Cost = path.Cost + int(link.IGPMetric) + int(previousLink.IGPMetric)

			pID := src + ":" + dst
			paths, ok := t.Paths[pID]
			if ok {
				paths = append(paths, path)
				t.Paths[pID] = paths
				return
			}
			t.Paths[pID] = []*Path{
				0: path,
			}
			return
		}
		if previousLink.RemoteNode == link.LocalNode {
			path.Links = append(path.Links, previousLink)
			path.Cost = path.Cost + int(previousLink.IGPMetric)
			t.FindPath(src, dst, link, path, getLinksCopyWithRemovedElement(links, i))
		}
	}
}

func (t *TopoView) HandleNodeNLRI(lsMessage *anypb.Any, p *api.Path) {
	defer t.Unlock()

	var NLRINode api.LsNodeNLRI
	err := ptypes.UnmarshalAny(lsMessage, &NLRINode)
	if err != nil {
		logrus.Println(err)
		return
	}
	node := &Node{
		ASN:        NLRINode.LocalNode.Asn,
		IGPRouteID: NLRINode.LocalNode.IgpRouterId,
	}
	var LsAttribute api.LsAttribute
	for _, item := range p.Pattrs {
		if ptypes.Is(item, &LsAttribute) {
			err := ptypes.UnmarshalAny(item, &LsAttribute)
			if err != nil {
				logrus.Println(err)
				return
			}
			node.RouterID = LsAttribute.Node.LocalRouterId
			node.Name = LsAttribute.Node.Name
			node.SRRangeStart = int(LsAttribute.Node.SrCapabilities.Ranges[0].Begin)
			node.SRRangeEnd = int(LsAttribute.Node.SrCapabilities.Ranges[0].End)
		}
	}
	t.Lock()
	t.NodesByIGPRouteID[node.IGPRouteID] = node
}

func (t *TopoView) HandleLinkNLRI(lsMessage *anypb.Any, p *api.Path) {
	defer t.Unlock()

	var NLRILink api.LsLinkNLRI
	err := ptypes.UnmarshalAny(lsMessage, &NLRILink)
	if err != nil {
		logrus.Println(err)
		return
	}
	link := &Link{
		LocalNode:   NLRILink.LocalNode.IgpRouterId,
		RemoteNode:  NLRILink.RemoteNode.IgpRouterId,
		IntIP:       NLRILink.LinkDescriptor.InterfaceAddrIpv4,
		NeighbourIP: NLRILink.LinkDescriptor.NeighborAddrIpv4,
	}
	var LsAttribute api.LsAttribute
	for _, item := range p.Pattrs {
		if ptypes.Is(item, &LsAttribute) {
			err := ptypes.UnmarshalAny(item, &LsAttribute)
			if err != nil {
				logrus.Println(err)
				return
			}
			link.BW = LsAttribute.Link.Bandwidth
			link.DefaultTEMetric = LsAttribute.Link.DefaultTeMetric
			link.IGPMetric = LsAttribute.Link.IgpMetric
			link.ReservableBW = LsAttribute.Link.ReservableBandwidth
			link.UnreservedBW = LsAttribute.Link.UnreservedBandwidth[0]
			link.SRAdjacencySID = LsAttribute.Link.SrAdjacencySid
		}
	}
	t.Lock()
	t.LinksByIGPRouteID = append(t.LinksByIGPRouteID, link)
}

func (t *TopoView) HandlePrefixV4NLRI(lsMessage *anypb.Any, p *api.Path) {
	defer t.Unlock()

	var NLRIPrefix api.LsPrefixV4NLRI
	err := ptypes.UnmarshalAny(lsMessage, &NLRIPrefix)
	if err != nil {
		logrus.Println(err)
		return
	}
	prefix := &Prefix{
		Prefix:    NLRIPrefix.PrefixDescriptor.IpReachability[0],
		LocalNode: NLRIPrefix.LocalNode.IgpRouterId,
	}
	var LsAttribute api.LsAttribute
	for _, item := range p.Pattrs {
		if ptypes.Is(item, &LsAttribute) {
			err := ptypes.UnmarshalAny(item, &LsAttribute)
			if err != nil {
				logrus.Println(err)
				return
			}
			prefix.SRPrefixSID = LsAttribute.Prefix.SrPrefixSid
		}
	}
	t.Lock()
	t.PrefixByIGPRouteID[prefix.LocalNode] = prefix
}

func (t *TopoView) Monitor(p *api.Path) {
	var lsMessage api.LsAddrPrefix
	err := ptypes.UnmarshalAny(p.Nlri, &lsMessage)
	if err != nil {
		logrus.Println(err)
		return
	}
	switch {
	case lsMessage.Type == 1:
		t.HandleNodeNLRI(lsMessage.Nlri, p)
	case lsMessage.Type == 2:
		t.HandleLinkNLRI(lsMessage.Nlri, p)
	case lsMessage.Type == 3:
		t.HandlePrefixV4NLRI(lsMessage.Nlri, p)
	default:
		logrus.Printf("uknown NLRI type: %d\n", lsMessage.Type)
		return
	}
	logrus.WithFields(logrus.Fields{
		"type":  "bgp",
		"event": "rcv_nlri",
		"nlri":  lsMessage,
	}).Info("recived NLRI")
	t.TopologyUpdate <- true
	logrus.WithFields(logrus.Fields{
		"type":  "bgp",
		"event": "sent_topo_update",
		"nlri":  lsMessage,
	}).Info("sent topology update into channel")
}

func (t *TopoView) findBestPath(bwNeeded int, src, dst string) *Path {
	var bestPath *Path
	for _, path := range t.Paths[src+":"+dst] {
		var bwAvailiable bool
		for _, link := range path.Links {
			if link.UnreservedBW > float32(bwNeeded) {
				bwAvailiable = true
				continue
			}
			bwAvailiable = false
		}
		if bwAvailiable && bestPath == nil {
			bestPath = path
			continue
		}
		if bwAvailiable && path.Cost < bestPath.Cost {
			bestPath = path
			continue
		}

	}
	return bestPath
}

func (t *TopoView) getSIDByIGPRouterID(routerID string) (uint32, error) {
	node, ok := t.NodesByIGPRouteID[routerID]
	if !ok {
		return 0, fmt.Errorf("no node found for id: %s", routerID)
	}
	prefix, ok := t.PrefixByIGPRouteID[routerID]
	if !ok {
		return 0, fmt.Errorf("no node found for id: %s", routerID)
	}
	return uint32(node.SRRangeStart) + prefix.SRPrefixSID, nil
}

func (t *TopoView) createSRLSP(bw uint32, path *Path) (*pcep.SRLSP, error) {
	defer t.Unlock()

	if len(path.Links) == 0 {
		return nil, fmt.Errorf("no links found in path")
	}

	t.Lock()
	srcPrefix, ok := t.PrefixByIGPRouteID[path.Src]
	if !ok {
		return nil, fmt.Errorf("src prefix not found for IGPID %s", path.Src)
	}
	dstPrefix, ok := t.PrefixByIGPRouteID[path.Dst]
	if !ok {
		return nil, fmt.Errorf("dst prefix not found for IGPID %s", path.Dst)
	}

	lspSrc := strings.Split(srcPrefix.Prefix, "/")[0]
	lspDst := strings.Split(dstPrefix.Prefix, "/")[0]

	lsp := &pcep.SRLSP{
		Delegate:     true,
		Sync:         false,
		Remove:       false,
		Admin:        true,
		Name:         "LSP-" + lspSrc + "-" + lspDst,
		Src:          lspSrc,
		Dst:          lspDst,
		SetupPrio:    7,
		HoldPrio:     7,
		LocalProtect: false,
		BW:           bw,
		EROList:      make([]pcep.SREROSub, 0),
	}

	for i, link := range path.Links {
		SID, err := t.getSIDByIGPRouterID(link.LocalNode)
		if err != nil {
			return nil, err
		}
		if i == 0 {
			lsp.EROList = append(lsp.EROList, pcep.SREROSub{
				LooseHop:   false,
				MBit:       true,
				NT:         3,
				IPv4NodeID: "",
				SID:        SID,
				NoSID:      false,
				IPv4Adjacency: []string{
					0: link.IntIP,
					1: link.NeighbourIP,
				},
			})
			continue
		}
		nodePrefix, ok := t.PrefixByIGPRouteID[link.LocalNode]
		if !ok {
			return nil, fmt.Errorf("node prefix not found for IGPID %s", link.LocalNode)
		}
		lsp.EROList = append(lsp.EROList, pcep.SREROSub{
			LooseHop:   false,
			MBit:       true,
			NT:         1,
			IPv4NodeID: strings.Split(nodePrefix.Prefix, "/")[0],
			SID:        SID,
			NoSID:      false,
		})
	}

	return lsp, nil
}

// used for debugging
// func readTopoFromFile() {
// 	data, err := ioutil.ReadFile("output.json")
// 	if err != nil {
// 		logrus.Error(err)
// 	}
// 	var topo TopoView

// 	err = json.Unmarshal(data, &topo)
// 	if err != nil {
// 		logrus.Error(err)
// 	}

// 	start := time.Now()
// 	for srcNode := range topo.NodesByIGPRouteID {
// 		for dstNode := range topo.NodesByIGPRouteID {
// 			if srcNode != dstNode {
// 				topo.FindAllPaths(srcNode, dstNode)
// 			}
// 		}
// 	}
// 	logrus.Printf("topo calc took %s", time.Since(start))

// 	// topo.FindAllPaths("0192.0168.0014", "0192.0168.0011")

// 	bestPath := topo.findBestPath(0, "0100.1001.0010", "0192.0168.0014")
// 	lsp, err := topo.createSRLSP(100, bestPath)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	printAsJSON(lsp)
// 	printAsJSON(topo.PrefixByIGPRouteID)
// }

func getPeers() []*api.Peer {
	return []*api.Peer{
		{
			Conf: &api.PeerConf{
				NeighborAddress: "10.0.0.10",
				PeerAs:          65000,
			},
			EbgpMultihop: &api.EbgpMultihop{
				Enabled:     true,
				MultihopTtl: 2,
			},
			ApplyPolicy: &api.ApplyPolicy{
				ImportPolicy: &api.PolicyAssignment{
					DefaultAction: api.RouteAction_ACCEPT,
				},
				ExportPolicy: &api.PolicyAssignment{
					DefaultAction: api.RouteAction_REJECT,
				},
			},
			AfiSafis: []*api.AfiSafi{
				{
					Config: &api.AfiSafiConfig{
						Family: &api.Family{
							Afi:  api.Family_AFI_LS,
							Safi: api.Family_SAFI_LS,
						},
						Enabled: true,
					},
				},
			},
		},
	}
}

type bgpGlobalCfg struct {
	AS         uint32
	RouterId   string
	ListenPort int32
}

func readBGPGlobalCfg() *bgpGlobalCfg {
	return &bgpGlobalCfg{
		AS:         65001,
		RouterId:   "19.19.19.19",
		ListenPort: -1, // gobgp won't listen on tcp:179
	}
}

func prepBGPStartRequest(cfg *bgpGlobalCfg) *api.StartBgpRequest {
	return &api.StartBgpRequest{
		Global: &api.Global{
			As:         cfg.AS,
			RouterId:   cfg.RouterId,
			ListenPort: cfg.ListenPort,
		},
	}
}

func (c *Controller) StartBGPLS() {

	c.bgpServer = gobgp.NewBgpServer()
	go c.bgpServer.Serve()

	bgpCfg := readBGPGlobalCfg()

	c.StopBGP = make(chan bool)

	err := c.bgpServer.StartBgp(context.Background(), prepBGPStartRequest(bgpCfg))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"type":  "bgp",
			"event": "start",
		}).Error(err)
	}

	err = c.bgpServer.MonitorPeer(context.Background(), &api.MonitorPeerRequest{}, func(p *api.Peer) { logrus.Info(p) })
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"type":  "bgp",
			"event": "start",
		}).Error("start MonitorPeer")
	}

	for _, peer := range getPeers() {
		err = c.bgpServer.AddPeer(context.Background(), &api.AddPeerRequest{Peer: peer})
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"type":  "bgp",
				"event": "start",
			}).Error("add peer")
		}
	}

	err = c.bgpServer.MonitorTable(context.Background(), &api.MonitorTableRequest{
		TableType: api.TableType_GLOBAL,
		Family: &api.Family{
			Afi:  api.Family_AFI_LS,
			Safi: api.Family_SAFI_LS,
		},
	}, func(p *api.Path) {
		c.TopoView.Monitor(p)
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"type":  "bgp",
			"event": "start",
		}).Error("start monitor table")
	}

	<-c.StopBGP

	err = c.bgpServer.StopBgp(context.Background(), &api.StopBgpRequest{})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"type":  "bgp",
			"event": "stop",
		}).Error(err)
	}
	logrus.WithFields(logrus.Fields{
		"type":  "bgp",
		"event": "bgp stop",
	}).Info("stopping bgp")

}

func (c *Controller) GetBGPNeighbor() ([]*api.Peer, error) {
	var list []*api.Peer

	err := c.bgpServer.ListPeer(context.Background(), &api.ListPeerRequest{}, func(n *api.Peer) {
		list = append(list, n)

	})
	if err != nil {
		return nil, err
	}
	return list, nil
}
