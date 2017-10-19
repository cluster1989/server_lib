package libzookeeper

type PartitionList []*Partition

func (pl PartitionList) Len() int {
	return len(pl)
}

func (pl PartitionList) Less(i, j int) bool {
	return pl[i].topic.Name < pl[j].topic.Name || (pl[i].topic.Name == pl[j].topic.Name && pl[i].ID < pl[j].ID)
}

func (pl PartitionList) Swap(i, j int) {
	pl[i], pl[j] = pl[j], pl[i]
}
