package metrics

func MergeLabels(m LabelSet, k []string, v []string) LabelSet {
	n := make(LabelSet)

	for k, v := range m {
		n[k] = v
	}

	for i, kv := range k {
		n[kv] = v[i]
	}
	return n
}

func MergeLabelsDirect(m LabelSet, o LabelSet) LabelSet {
	n := make(LabelSet)

	for k, v := range m {
		n[k] = v
	}

	for k, v := range o {
		n[k] = v
	}
	return n
}
