package define

type SubMsgStruct struct {
	IpLink           string
	IpLinkUpdateTime string

	DomainLink           string
	DomainLinkUpdateTime string
}

type ConfigKey struct {
	IpSubLink               string
	IpSubLinkUpdateTime     string
	DomainSubLink           string
	DomainSubLinkUpdateTime string
	ApiKey                  string
	CornInterval            string
}

type SystemHealthMsg struct {
	IpSubLinkStatus         bool
	IpSubLinkUpdateTime     string
	DomainSubLinkStatus     bool
	DomainSubLinkUpdateTime string

	LocalIpSubLinkStatus         bool
	LocalIpSubLinkUpdateTime     string
	LocalDomainSubLinkStatus     bool
	LocalDomainSubLinkUpdateTime string
}

type SubLinkType string

const JmsMirrorListFileName = ".conf/jms-mirror"
const JmsSubApi = "/members/getsub.php"
