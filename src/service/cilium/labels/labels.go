package labels

import (
	"fmt"
	"sort"
	"strings"
)

var (
	LabelSourceKVM = "kvm"
)

const (
	PolicyLabelName        = "io.kubearmor.kvm.policy.name"
	PolicyLabelUID         = "io.kubearmor.kvm.policy.uid"
	PolicyLabelDerivedFrom = "io.kubearmor.kvm.policy.derived-from"
)

type Label struct {
	Key    string `json:"key"`
	Value  string `json:"value,omitempty"`
	Source string `json:"source"`
}

type LabelArray []Label

type LabelMap map[string]Label

func NewLabel(key string, value string, source string) Label {
	return Label{
		Key:    key,
		Value:  value,
		Source: source,
	}
}

func NewLabelMap(labels []string) LabelMap {
	lblMap := make(LabelMap, len(labels))

	for _, label := range labels {
		lbl := ParseLabel(label)
		lblMap[lbl.Key] = lbl
	}

	return lblMap
}

func ParseLabel(str string) (lbl Label) {
	i := strings.IndexByte(str, ':')
	if i < 0 {
		lbl.Source = LabelSourceKVM
	} else if i == 0 {
		lbl.Source = LabelSourceKVM
		str = str[i+1:]
	} else {
		lbl.Source = str[:i]
		str = str[i+1:]
	}

	i = strings.IndexByte(str, '=')
	if i < 0 {
		lbl.Key = str
	} else if i == 0 && lbl.Source == LabelSourceKVM {
		lbl.Key = str[i+1:]
	} else {
		lbl.Key = str[:i]
		lbl.Value = str[i+1:]
	}

	return lbl
}

func (l Label) formatForKVStore() string {
	return fmt.Sprintf(`%s:%s=%s;`, l.Source, l.Key, l.Value)
}

func (l LabelMap) SortedList() string {
	keys := make([]string, 0, len(l))
	for k := range l {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	result := ""
	for _, k := range keys {
		result += l[k].formatForKVStore()
	}

	return result
}
