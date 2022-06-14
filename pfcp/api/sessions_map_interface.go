package api

type SessionsMapInterface interface {
	Add(session PFCPSessionInterface) error
	GetPFCPSessions() []PFCPSessionInterface
	GetPFCPSession(localIP string, seid SEID) (PFCPSessionInterface, error)
	Update(session PFCPSessionInterface) error
}
