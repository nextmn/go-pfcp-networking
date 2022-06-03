# go-pfcp-networking: PFCP Networking functionnalities on top of go-pfcp 

Still a Work In Progress. API may change before v1.0.0.

## Features
- PFCP Sessions handling (currently only PFCP Session establishment procedure is supported)

## Getting started
### UPF

```golang
upNode := NewPFCPEntityUP(UPFADDR)
upNode.Start()
// Access list of associations
associations := upNode.GetPFCPAssociations()
// Access list of sessions
sessions := upNode.GetPFCPSessions()
```

### SMF

```golang
cpNode := NewPFCPEntityCP(SMFADDR)
cpNode.Start()
association, _ := cpNode.NewEstablishedAssociation(ie.NewNodeIDHeuristic(UPFADDR))
session, _ := a.CreateSession(pdrs, fars)

```

## Author
Louis Royer

## License
MIT
