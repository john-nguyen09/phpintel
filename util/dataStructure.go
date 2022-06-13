package util

import "github.com/zyedidia/generic/mapset"

func SetFromArray[V comparable](arr []V) mapset.Set[V] {
	set := mapset.New[V]()
	for _, el := range arr {
		set.Put(el)
	}
	return set
}
