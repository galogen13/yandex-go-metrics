package server

type AuditServise struct {
	Auditors []Auditor
}

func NewAuditServise() *AuditServise {
	return &AuditServise{Auditors: []Auditor{}}
}

func (as *AuditServise) Register(auditor Auditor) {
	as.Auditors = append(as.Auditors, auditor)
}

func (as *AuditServise) Notify(auditLog AuditLog) {
	for _, auditor := range as.Auditors {
		go auditor.Notify(auditLog)
	}
}

type AuditLog struct {
	ts        int64    `json:"ts"`
	metrics   []string `json:"metrics"`
	ipAddress string   `json:"ip_address"`
}

type Auditor interface {
	Notify(AuditLog)
}

type URLAuditor struct {
	url string
}

func (urlAuditor URLAuditor) Notify(auditLog AuditLog) {

}

func NewURLAuditor(url string) *URLAuditor {
	return &URLAuditor{url: url}
}

type FileAuditor struct {
	filePath string
}

func NewFileAuditor(filePath string) *FileAuditor {
	return &FileAuditor{filePath: filePath}
}

func (fileAuditor FileAuditor) Notify(auditLog AuditLog) {

}
