package postbar

type Content interface{}

type Pic struct {
	Src string
}

type Video struct {
	Src string
}

type Music struct {
	Src string
}

type Text struct {
	Text string
}

type Emoticon struct {
	Text string
	C    string //描述
}

type Voice struct {
	DuringTime int32
	VoiceMD5   string //voice_md5
}

type Link struct {
	Link string
	Text string
}

type Packet struct {
	C          string
	PacketName string
}

type At struct {
	Text string
	Uid  uint64
}
