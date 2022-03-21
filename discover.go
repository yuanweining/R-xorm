package Rxorm

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type Discovery struct{
	hash Hash
	keys []int
	replicas int // 一个真实对应的虚拟节点数
	hashMap map[int]string // 虚拟 -> 真实
}

func NewDiscovery(replicas int, fn Hash) *Discovery {
	m := &Discovery{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// 输入真实机器地址
func(d *Discovery) Add(addrs ...string){
	for _, addr := range addrs{
		for i:=0;i<d.replicas;i++{
			hash := int(d.hash([]byte(strconv.Itoa(i)+addr)))
			d.keys = append(d.keys, hash)
			d.hashMap[hash] = addr
		}
	}
	sort.Ints(d.keys)
}

// 输入key，返回真值机器地址
func(d *Discovery) Get(key string) string{
	if len(d.keys) == 0 {
		return ""
	}
	hash := int(d.hash([]byte(key)))
	idx := sort.Search(len(d.keys), func(i int) bool { //希望前面为假，后面为真，返回第一个真值
		return d.keys[i] >= hash
	})
	return d.hashMap[d.keys[idx%len(d.keys)]]
}

