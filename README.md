# go-pfcp-networking: PFCP Networking functionnalities on top of go-pfcp 

Still a Work In Progress. API may change before v1.0.0.

## Features
- PFCP Sessions handling (currently only PFCP Session establishment procedure is supported)

## Getting started
### Server (UPF)

```golang
upNode := NewPFCPServerEntity(upAddress)
upNode.Start()
// Access list of sessions
sessions := upNode.GetPFCPSessions()
```

### Client (SMF)

```golang
cpNode := NewPFCPClientEntity(cpAddress)
cpNode.Start()
peer, _ := NewPFCPPeer(cpNode, pfcputils.CreateNodeID(nodeID)
a, _ := cpNode.NewPFCPAssociation(peer)
a.NewPFCPSession(cpNode.GetNextRemoteSessionID(), pdrs, fars)

```

## Author
Louis Royer

## License
MIT
