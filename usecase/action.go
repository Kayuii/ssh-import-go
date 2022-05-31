package usecase

type CmdProtocol int8
type GitProto int8

const (
	CMD_DEFAULT CmdProtocol = iota
	CMD_COLOR
	CMD_OUTPUT
	CMD_REMOVE
	CMD_USERAGENT
)

const (
	DEFAULT_PROTO GitProto = iota
	GIT_GITHUB    GitProto = iota - 1
	GIT_LAUNCHPAD
)

var GitProtoMap = map[GitProto]string{
	GIT_GITHUB:    "gh",
	GIT_LAUNCHPAD: "lp",
}

func (g GitProto) String() string {
	return GitProtoMap[g]
}

func (g GitProto) Is(proto string) GitProto {
	for k, v := range GitProtoMap {
		if v == proto {
			return k
		}
	}
	return DEFAULT_PROTO
}

func (g GitProto) Check(proto string) bool {
	for k, v := range GitProtoMap {
		if k == g && v == proto {
			return true
		}
	}
	return false
}

func (g GitProto) Match(proto string) bool {
	for _, v := range GitProtoMap {
		if v == proto {
			return true
		}
	}
	return false
}
