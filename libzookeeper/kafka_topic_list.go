package libzookeeper

type TopicList []*Topic

func (tl TopicList) Find(name string) *Topic {
	for _, topic := range tl {
		if topic.Name == name {
			return topic
		}
	}
	return nil
}

func (tl TopicList) Len() int {
	return len(tl)
}

func (tl TopicList) Less(i, j int) bool {
	return tl[i].Name < tl[j].Name
}

func (tl TopicList) Swap(i, j int) {
	tl[i], tl[j] = tl[j], tl[i]
}
