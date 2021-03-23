package xmind

type Content []Sheet

type Sheet struct {
	Topic Topic `json:"topic,omitempty" xml:"topic,omitempty"`
}

type Topic struct {
	Title  string  `json:"title,omitempty" xml:"title,omitempty"`
	Topics []Topic `json:"topics,omitempty" xml:"topics,omitempty"`
}

func (t Topic) GetFirstSubTopicTitle() string {
	if len(t.Topics) == 0 {
		return ""
	}
	return t.Topics[0].Title
}
