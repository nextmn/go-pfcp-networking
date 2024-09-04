# go-pfcp-networking: PFCP Networking functionalities on top of go-pfcp

> [!WARNING]
> Still a Work In Progress. API may change before v1.0.0.

## Features
- PFCP Sessions handling (currently only PFCP Session establishment procedure is supported)

## Getting started
### UPF

```golang
upNode := NewPFCPEntityUP(UPF_NODE_ID, UPF_IP_ADDR) // node id can be an IP Address or a FQDN
go upNode.ListenAndServe()
defer upNode.Close()
// Access list of associations
associations := upNode.GetPFCPAssociations()
// Access list of sessions
sessions := upNode.GetPFCPSessions()
```

### SMF

```golang
cpNode := NewPFCPEntityCP(SMF_NODE_ID, SMF_IP_ADDR) // node id can be an IP Address or a FQDN
go cpNode.ListenAndServe()
defer cpNode.Close()
association, _ := cpNode.NewEstablishedPFCPAssociation(ie.NewNodeIDHeuristic(UPFADDR))
session, _ := a.CreateSession(pdrs, fars)

```

## Author
Louis Royer

## License
MIT
