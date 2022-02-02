package pfcp_networking

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/louisroyer/go-pfcp-networking/pfcputil"
	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

type messageChan chan []byte

type PFCPPeer struct {
	NodeID  *ie.IE
	Srv     PFCPEntityInterface
	conn    *net.UDPConn
	udpAddr *net.UDPAddr
	seq     uint32
	queue   map[uint32]messageChan
	mu      sync.Mutex
	stop    bool
}

func NewPFCPPeer(srv PFCPEntityInterface, nodeID *ie.IE) (peer *PFCPPeer, err error) {
	ipAddr, err := nodeID.NodeID()
	if err != nil {
		return nil, err
	}
	udpAddr := pfcputil.CreateUDPAddr(ipAddr, pfcputil.PFCP_PORT)
	raddr, err := net.ResolveUDPAddr("udp", udpAddr)
	if err != nil {
		return nil, err
	}
	c, err := net.Dial("udp", udpAddr)
	if err != nil {
		return nil, err
	}
	c.Close()
	laddr := c.LocalAddr().(*net.UDPAddr)
	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return nil, err
	}
	p := PFCPPeer{
		Srv:     srv,
		NodeID:  nodeID,
		conn:    conn,
		udpAddr: raddr,
		seq:     1,
		queue:   make(map[uint32]messageChan),
		mu:      sync.Mutex{},
		stop:    false,
	}
	// Read incomming messages
	go func(e PFCPPeer) {
		var stop bool = false
		for !stop {
			b := make([]byte, pfcputil.DEFAULT_MTU) // TODO: detect MTU for interface instead of using DEFAULT_MTU
			n, _, err := e.conn.ReadFromUDP(b)
			if err != nil {
				// socket has been closed
				return
			}
			msg, err := message.ParseHeader(b[:n])
			if err != nil {
				e.mu.Lock()
				stop = e.stop
				e.mu.Unlock()
				if stop {
					return
				} else {
					continue
				}
			}
			sn := msg.SequenceNumber
			e.mu.Lock()
			ch, exists := e.queue[sn]
			if exists {
				ch <- b[:n]
			}
			stop = e.stop
			e.mu.Unlock()
		}
	}(p)
	return &p, nil
}

// Close connection of PFCPPeer
func (peer *PFCPPeer) Close() error {
	peer.mu.Lock()
	if !peer.stop {
		peer.stop = true
		err := peer.conn.Close()
		if err != nil {
			return err
		}
	}
	peer.mu.Unlock()
	return nil
}

// Get next sequence number available for this PFCPPeer
func (peer *PFCPPeer) getNextSequenceNumber() uint32 {
	peer.mu.Lock()
	s := peer.seq
	peer.seq += 1
	peer.mu.Unlock()
	return s
}

// Add a message to queue. Response will be send to channel ch messageChan
func (peer *PFCPPeer) addToQueue(sn uint32, ch messageChan) {
	peer.mu.Lock()
	peer.queue[sn] = ch
	peer.mu.Unlock()
}

// Remove a message from queue (used when a response is received, or when timeout is reached)
func (peer *PFCPPeer) deleteFromQueue(sn uint32) {
	peer.mu.Lock()
	close(peer.queue[sn])
	delete(peer.queue, sn)
	peer.mu.Unlock()
}

// Send a PFCP message
func (peer *PFCPPeer) Send(msg message.Message) (m message.Message, err error) {
	// timer
	T1 := time.Millisecond * 500
	// retries
	N1 := 3

	//XXX: cannot use `h, err := msg.(*message.Header)` because Header does not implement MessageTypeName()
	msgb := make([]byte, msg.MarshalLen())
	err = msg.MarshalTo(msgb)
	if err != nil {
		return nil, err
	}
	h, err := message.ParseHeader(msgb)
	if err != nil {
		return nil, err
	}

	if !pfcputil.IsMessageTypeRequest(h.MessageType()) {
		return nil, fmt.Errorf("Unexpected outcomming PFCP message type")
	}
	sn := peer.getNextSequenceNumber()
	h.SetSequenceNumber(sn)
	b, err := h.Marshal()
	if err != nil {
		return nil, err
	}

	ch := make(messageChan)
	peer.addToQueue(sn, ch)
	defer peer.deleteFromQueue(sn)

	_, err = peer.conn.WriteToUDP(b, peer.udpAddr)
	if err != nil {
		return nil, fmt.Errorf("Error on write: %s\n", err)
	}

	for i := 0; i < N1; i++ {
		select {
		case r := <-ch:
			msg, err := message.Parse(r)
			if err != nil {
				return nil, fmt.Errorf("Unexpected incomming packet")
			}
			if !pfcputil.IsMessageTypeResponse(msg.MessageType()) {
				return nil, fmt.Errorf("Unexpected incomming PFCP message type")
			}
			return msg, nil
		case <-time.After(T1):
			// retry
			_, err = peer.conn.WriteToUDP(b, peer.udpAddr)
			if err != nil {
				return nil, fmt.Errorf("Error on write: %s\n", err)
			}
		}
	}
	return nil, fmt.Errorf("Unsucessfull transfer of Request message")
}

// Send an Heartbeat request, return true if the PFCP peer is alive.
func (peer *PFCPPeer) IsAlive() (res bool, err error) {
	if peer.Srv.RecoveryTimeStamp() == nil {
		return false, fmt.Errorf("SMF is not started.")
	}
	hreq := message.NewHeartbeatRequest(
		0,
		peer.Srv.RecoveryTimeStamp(),
		nil)

	_, err = peer.Send(hreq)
	if err != nil {
		return false, err
	}
	return true, nil
}
