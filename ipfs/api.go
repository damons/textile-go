package ipfs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"sort"
	"strings"
	"time"

	"gx/ipfs/QmPSQnBKM9g7BaUcZCvswUJVscQ1ipjmwxN5PXCjkp9EQ7/go-cid"
	libp2pc "gx/ipfs/QmPvyPwuCgJ7pDmrKDxRtsScJgBaM5h4EpRL2qQJsmXf4n/go-libp2p-crypto"
	ipld "gx/ipfs/QmR7TcHkR9nxkUorfi8XMTAMLUK7GiP64TWWBzY3aacc1o/go-ipld-format"
	"gx/ipfs/QmT3rzed1ppXefourpmoZ7tyVQfsGPQZ1pHDngLmCvXxd3/go-path"
	"gx/ipfs/QmTRhk7cgjUf2gfQ3p2M9KPECNZEW9XUrmHcFCgog4cPgB/go-libp2p-peer"
	"gx/ipfs/QmUJYo4etAQqFfSS2rarFAE97eNGB8ej64YkRT2SmsYD4r/go-ipfs/core"
	"gx/ipfs/QmUJYo4etAQqFfSS2rarFAE97eNGB8ej64YkRT2SmsYD4r/go-ipfs/core/coreapi"
	"gx/ipfs/QmUJYo4etAQqFfSS2rarFAE97eNGB8ej64YkRT2SmsYD4r/go-ipfs/core/coreapi/interface"
	"gx/ipfs/QmUJYo4etAQqFfSS2rarFAE97eNGB8ej64YkRT2SmsYD4r/go-ipfs/core/coreapi/interface/options"
	"gx/ipfs/QmUJYo4etAQqFfSS2rarFAE97eNGB8ej64YkRT2SmsYD4r/go-ipfs/core/coreunix"
	"gx/ipfs/QmUJYo4etAQqFfSS2rarFAE97eNGB8ej64YkRT2SmsYD4r/go-ipfs/namesys/opts"
	"gx/ipfs/QmUJYo4etAQqFfSS2rarFAE97eNGB8ej64YkRT2SmsYD4r/go-ipfs/pin"
	inet "gx/ipfs/QmXuRkCR7BNQa9uqfpTiFWsTQLzmTWYg91Ja1w95gnqb6u/go-libp2p-net"
	logging "gx/ipfs/QmZChCsSt8DctjceaL56Eibc29CVQq4dGKRXC5JRZ6Ppae/go-log"
	"gx/ipfs/QmZMWMvWMVKCbHetJ4RgndbuEF1io2UpUxwQwtNjtYPzSC/go-ipfs-files"
	uio "gx/ipfs/QmfB3oNXGGq9S4B2a9YeCajoATms3Zw2VvDm8fK7VeLSV8/go-unixfs/io"
)

var log = logging.Logger("tex-ipfs")

const pinTimeout = time.Minute
const catTimeout = time.Second * 30
const ipnsTimeout = time.Second * 10
const connectTimeout = time.Second * 10

type IpnsEntry struct {
	Name  string
	Value string
}

// DataAtPath return bytes under an ipfs path
func DataAtPath(node *core.IpfsNode, pth string) ([]byte, error) {
	ip, err := iface.ParsePath(pth)
	if err != nil {
		return nil, err
	}

	api := coreapi.NewCoreAPI(node)
	ctx, cancel := context.WithTimeout(node.Context(), catTimeout)
	defer cancel()

	reader, err := api.Unixfs().Get(ctx, ip)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return ioutil.ReadAll(reader)
}

// LinksAtPath return ipld links under a path
func LinksAtPath(node *core.IpfsNode, pth string) ([]*ipld.Link, error) {
	ip, err := iface.ParsePath(pth)
	if err != nil {
		return nil, err
	}

	api := coreapi.NewCoreAPI(node)
	ctx, cancel := context.WithTimeout(node.Context(), catTimeout)
	defer cancel()

	links, err := api.Unixfs().Ls(ctx, ip)
	if err != nil {
		return nil, err
	}

	return links, nil
}

// AddDataToDirectory adds reader bytes to a virtual dir
func AddDataToDirectory(node *core.IpfsNode, dir uio.Directory, fname string, reader io.Reader) (*cid.Cid, error) {
	str, err := coreunix.Add(node, reader)
	if err != nil {
		return nil, err
	}

	id, err := cid.Decode(str)
	if err != nil {
		return nil, err
	}

	n, err := node.DAG.Get(node.Context(), id)
	if err != nil {
		return nil, err
	}

	if err := dir.AddChild(node.Context(), fname, n); err != nil {
		return nil, err
	}

	return &id, nil
}

// AddLinkToDirectory adds a link to a virtual dir
func AddLinkToDirectory(node *core.IpfsNode, dir uio.Directory, fname string, pth string) error {
	id, err := cid.Decode(pth)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(node.Context(), catTimeout)
	defer cancel()

	nd, err := node.DAG.Get(ctx, id)
	if err != nil {
		return err
	}

	ctx2, cancel2 := context.WithTimeout(node.Context(), catTimeout)
	defer cancel2()

	return dir.AddChild(ctx2, fname, nd)
}

// AddData takes a reader and adds it, optionally pins it
func AddData(node *core.IpfsNode, reader io.Reader, pin bool) (*cid.Cid, error) {
	ctx, cancel := context.WithTimeout(node.Context(), pinTimeout)
	defer cancel()

	api := coreapi.NewCoreAPI(node)
	pth, err := api.Unixfs().Add(ctx, dataFile(reader)())
	if err != nil {
		return nil, err
	}

	if pin {
		if err := api.Pin().Add(ctx, pth); err != nil {
			return nil, err
		}
	}
	id := pth.Cid()

	return &id, nil
}

// PinPath takes an ipfs path string and pins it
func PinPath(node *core.IpfsNode, pth string, recursive bool) error {
	ip, err := iface.ParsePath(pth)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(node.Context(), pinTimeout)
	defer cancel()

	api := coreapi.NewCoreAPI(node)
	if err := api.Pin().Add(ctx, ip, options.Pin.Recursive(recursive)); err != nil {
		return err
	}

	return nil
}

// UnpinPath takes an ipfs path string and unpins it
// Note: This is always recursive. Use UnpinNode for finer control.
func UnpinPath(node *core.IpfsNode, pth string) error {
	ip, err := iface.ParsePath(pth)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(node.Context(), pinTimeout)
	defer cancel()

	api := coreapi.NewCoreAPI(node)
	if err := api.Pin().Rm(ctx, ip); err != nil && err != pin.ErrNotPinned {
		return err
	}

	return nil
}

// NodeAtLink returns the node behind an ipld link
func NodeAtLink(node *core.IpfsNode, link *ipld.Link) (ipld.Node, error) {
	ctx, cancel := context.WithTimeout(node.Context(), catTimeout)
	defer cancel()
	return link.GetNode(ctx, node.DAG)
}

// NodeAtCid returns the node behind a cid
func NodeAtCid(node *core.IpfsNode, id *cid.Cid) (ipld.Node, error) {
	ctx, cancel := context.WithTimeout(node.Context(), catTimeout)
	defer cancel()
	return node.DAG.Get(ctx, *id)
}

// NodeAtPath returns the last node under path
func NodeAtPath(node *core.IpfsNode, pth string) (ipld.Node, error) {
	p, err := path.ParsePath(pth)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(node.Context(), catTimeout)
	defer cancel()

	return node.Resolver.ResolvePath(ctx, p)
}

// PinNode pins an ipld node
func PinNode(node *core.IpfsNode, nd ipld.Node, recursive bool) error {
	ctx, cancel := context.WithTimeout(node.Context(), pinTimeout)
	defer cancel()

	if err := node.Pinning.Pin(ctx, nd, recursive); err != nil {
		if strings.Contains(err.Error(), "already pinned recursively") {
			return nil
		}
		return err
	}

	return node.Pinning.Flush()
}

// UnpinNode unpins an ipld node
func UnpinNode(node *core.IpfsNode, nd ipld.Node, recursive bool) error {
	ctx, cancel := context.WithTimeout(node.Context(), pinTimeout)
	defer cancel()

	err := node.Pinning.Unpin(ctx, nd.Cid(), recursive)
	if err != nil && err != pin.ErrNotPinned {
		return err
	}

	return node.Pinning.Flush()
}

// Publish publishes a content id to ipns
func Publish(node *core.IpfsNode, sk libp2pc.PrivKey, id string, dur time.Duration, cache time.Duration) (*IpnsEntry, error) {
	if node.Mounts.Ipns != nil && node.Mounts.Ipns.IsActive() {
		return nil, errors.New("cannot manually publish while IPNS is mounted")
	}

	pth, err := path.ParsePath(id)
	if err != nil {
		return nil, err
	}

	eol := time.Now().Add(dur)
	ctx, cancel := context.WithTimeout(node.Context(), ipnsTimeout)
	ctx = context.WithValue(ctx, "ipns-publish-ttl", cache)
	defer cancel()

	if err := node.Namesys.PublishWithEOL(ctx, sk, pth, eol); err != nil {
		return nil, err
	}

	pid, err := peer.IDFromPrivateKey(sk)
	if err != nil {
		return nil, err
	}

	return &IpnsEntry{Name: pid.Pretty(), Value: pth.String()}, nil
}

// Resolve resolves an ipns path to an ipfs path
func Resolve(node *core.IpfsNode, name peer.ID) (*path.Path, error) {
	key := fmt.Sprintf("/ipns/%s", name.Pretty())

	var ropts []nsopts.ResolveOpt
	ropts = append(ropts, nsopts.Depth(1))
	ropts = append(ropts, nsopts.DhtRecordCount(16))
	ropts = append(ropts, nsopts.DhtTimeout(ipnsTimeout))

	pth, err := node.Namesys.Resolve(node.Context(), key, ropts...)
	if err != nil {
		return nil, err
	}

	return &pth, nil
}

// SwarmConnect opens a direct connection to a list of peer multi addresses
func SwarmConnect(node *core.IpfsNode, addrs []string) ([]string, error) {
	pis, err := peersWithAddresses(addrs)
	if err != nil {
		return nil, err
	}

	api := coreapi.NewCoreAPI(node)
	ctx, cancel := context.WithTimeout(node.Context(), connectTimeout)
	defer cancel()

	output := make([]string, len(pis))
	for i, pi := range pis {
		output[i] = "connect " + pi.ID.Pretty()

		err := api.Swarm().Connect(ctx, pi)
		if err != nil {
			return nil, fmt.Errorf("%s failure: %s", output[i], err)
		}
		output[i] += " success"
	}

	return output, nil
}

type streamInfo struct {
	Protocol string
}

type connInfo struct {
	Addr      string         `json:"addr"`
	Peer      string         `json:"peer"`
	Latency   string         `json:"latency,omitempty"`
	Muxer     string         `json:"muxer,omitempty"`
	Direction inet.Direction `json:"direction,omitempty"`
	Streams   []streamInfo   `json:"streams,omitempty"`
}

func (ci *connInfo) Less(i, j int) bool {
	return ci.Streams[i].Protocol < ci.Streams[j].Protocol
}

func (ci *connInfo) Len() int {
	return len(ci.Streams)
}

func (ci *connInfo) Swap(i, j int) {
	ci.Streams[i], ci.Streams[j] = ci.Streams[j], ci.Streams[i]
}

type ConnInfos struct {
	Peers []connInfo
}

func (ci ConnInfos) Less(i, j int) bool {
	return ci.Peers[i].Addr < ci.Peers[j].Addr
}

func (ci ConnInfos) Len() int {
	return len(ci.Peers)
}

func (ci ConnInfos) Swap(i, j int) {
	ci.Peers[i], ci.Peers[j] = ci.Peers[j], ci.Peers[i]
}

// SwarmPeers lists the set of peers this node is connected to
func SwarmPeers(node *core.IpfsNode, verbose bool, latency bool, streams bool, direction bool) (*ConnInfos, error) {
	api := coreapi.NewCoreAPI(node)
	ctx, cancel := context.WithTimeout(node.Context(), connectTimeout)
	defer cancel()

	conns, err := api.Swarm().Peers(ctx)
	if err != nil {
		return nil, err
	}

	var out ConnInfos
	for _, c := range conns {
		ci := connInfo{
			Addr: c.Address().String(),
			Peer: c.ID().Pretty(),
		}

		if verbose || direction {
			// set direction
			ci.Direction = c.Direction()
		}

		if verbose || latency {
			lat, err := c.Latency()
			if err != nil {
				return nil, err
			}

			if lat == 0 {
				ci.Latency = "n/a"
			} else {
				ci.Latency = lat.String()
			}
		}
		if verbose || streams {
			strs, err := c.Streams()
			if err != nil {
				return nil, err
			}

			for _, s := range strs {
				ci.Streams = append(ci.Streams, streamInfo{Protocol: string(s)})
			}
		}
		sort.Sort(&ci)
		out.Peers = append(out.Peers, ci)
	}

	sort.Sort(&out)
	return &out, nil
}

func dataFile(reader io.Reader) func() files.File {
	return func() files.File {
		return files.NewReaderFile("", "", ioutil.NopCloser(reader), nil)
	}
}
