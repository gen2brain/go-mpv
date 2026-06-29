package mpv

import (
	"math"
	"unsafe"
)

// cAlloc and cFree are the C heap allocators, set by the active backend.
var (
	cAlloc func(size int) unsafe.Pointer
	cFree  func(p unsafe.Pointer)
)

// cNode, cNodeList and cByteArray mirror the C mpv_node types.
type cNode struct {
	u      uint64 // value, or pointer in its low bits
	format int32
}

type cNodeList struct {
	num    int32
	values unsafe.Pointer // *cNode
	keys   unsafe.Pointer // array of C strings
}

type cByteArray struct {
	data unsafe.Pointer
	size uintptr
}

var (
	nodeSize = unsafe.Sizeof(cNode{})
	listSize = unsafe.Sizeof(cNodeList{})
	baSize   = unsafe.Sizeof(cByteArray{})
	ptrSize  = unsafe.Sizeof(uintptr(0))
)

// nodePtr reinterprets the node union as a pointer.
func nodePtr(n *cNode) unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(&n.u))
}

// nodeToGo converts an mpv_node at p into a Go value.
func nodeToGo(p unsafe.Pointer) any {
	n := (*cNode)(p)

	switch Format(n.format) {
	case FormatString:
		return toStr(nodePtr(n))
	case FormatFlag:
		return uint32(n.u) != 0
	case FormatInt64:
		return int64(n.u)
	case FormatDouble:
		return math.Float64frombits(n.u)
	case FormatNodeArray:
		return nodeListToSlice(nodePtr(n))
	case FormatNodeMap:
		return nodeListToMap(nodePtr(n))
	case FormatByteArray:
		return byteArrayToGo(nodePtr(n))
	default:
		return nil
	}
}

func nodeListToSlice(p unsafe.Pointer) []any {
	if p == nil {
		return nil
	}

	list := (*cNodeList)(p)
	out := make([]any, list.num)
	if list.values != nil {
		nodes := unsafe.Slice((*cNode)(list.values), int(list.num))
		for i := range nodes {
			out[i] = nodeToGo(unsafe.Pointer(&nodes[i]))
		}
	}

	return out
}

func nodeListToMap(p unsafe.Pointer) map[string]any {
	if p == nil {
		return nil
	}

	list := (*cNodeList)(p)
	out := make(map[string]any, list.num)
	if list.values != nil && list.keys != nil {
		nodes := unsafe.Slice((*cNode)(list.values), int(list.num))
		keys := unsafe.Slice((*unsafe.Pointer)(list.keys), int(list.num))
		for i := range nodes {
			out[toStr(keys[i])] = nodeToGo(unsafe.Pointer(&nodes[i]))
		}
	}

	return out
}

func byteArrayToGo(p unsafe.Pointer) []byte {
	if p == nil {
		return nil
	}

	ba := (*cByteArray)(p)
	if ba.data == nil || ba.size == 0 {
		return []byte{}
	}

	out := make([]byte, ba.size)
	copy(out, unsafe.Slice((*byte)(ba.data), int(ba.size)))

	return out
}

// goToNode builds an mpv_node tree in C heap from v, returning the root pointer
// and a cleanup function that frees the whole tree.
func goToNode(v any) (unsafe.Pointer, func()) {
	root := (*cNode)(cAlloc(int(nodeSize)))
	fillNode(root, v)

	return unsafe.Pointer(root), func() {
		freeNode(root)
		cFree(unsafe.Pointer(root))
	}
}

// fillNode populates the allocated node dst from v, allocating referenced memory.
func fillNode(dst *cNode, v any) {
	dst.u = 0

	switch val := v.(type) {
	case nil:
		dst.format = int32(FormatNone)
	case string:
		dst.format = int32(FormatString)
		dst.u = uint64(uintptr(cString(val)))
	case bool:
		dst.format = int32(FormatFlag)
		if val {
			dst.u = 1
		}
	case int:
		dst.format = int32(FormatInt64)
		dst.u = uint64(int64(val))
	case int64:
		dst.format = int32(FormatInt64)
		dst.u = uint64(val)
	case float64:
		dst.format = int32(FormatDouble)
		dst.u = math.Float64bits(val)
	case []byte:
		dst.format = int32(FormatByteArray)
		dst.u = uint64(uintptr(buildByteArray(val)))
	case []any:
		dst.format = int32(FormatNodeArray)
		dst.u = uint64(uintptr(buildList(val, nil)))
	case map[string]any:
		dst.format = int32(FormatNodeMap)
		dst.u = uint64(uintptr(buildMap(val)))
	default:
		dst.format = int32(FormatNone)
	}
}

// freeNode releases the C memory referenced by n. It does not free n itself.
func freeNode(n *cNode) {
	switch Format(n.format) {
	case FormatString:
		cFree(nodePtr(n))
	case FormatByteArray:
		ba := (*cByteArray)(nodePtr(n))
		if ba.data != nil {
			cFree(ba.data)
		}
		cFree(unsafe.Pointer(ba))
	case FormatNodeArray, FormatNodeMap:
		list := (*cNodeList)(nodePtr(n))
		if list.values != nil {
			nodes := unsafe.Slice((*cNode)(list.values), int(list.num))
			for i := range nodes {
				freeNode(&nodes[i])
			}
			cFree(list.values)
		}
		if list.keys != nil {
			keys := unsafe.Slice((*unsafe.Pointer)(list.keys), int(list.num))
			for i := range keys {
				cFree(keys[i])
			}
			cFree(list.keys)
		}
		cFree(unsafe.Pointer(list))
	}
}

func buildByteArray(b []byte) unsafe.Pointer {
	ba := (*cByteArray)(cAlloc(int(baSize)))
	ba.data = nil
	ba.size = uintptr(len(b))
	if len(b) > 0 {
		data := cAlloc(len(b))
		copy(unsafe.Slice((*byte)(data), len(b)), b)
		ba.data = data
	}

	return unsafe.Pointer(ba)
}

// buildList builds a NODE_ARRAY when keys is nil, otherwise a NODE_MAP with
// keys[i] belonging to values[i].
func buildList(values []any, keys []string) unsafe.Pointer {
	list := (*cNodeList)(cAlloc(int(listSize)))
	list.num = int32(len(values))
	list.values = nil
	list.keys = nil
	if len(values) == 0 {
		return unsafe.Pointer(list)
	}

	nodes := cAlloc(len(values) * int(nodeSize))
	list.values = nodes
	dst := unsafe.Slice((*cNode)(nodes), len(values))
	for i := range values {
		fillNode(&dst[i], values[i])
	}

	if keys != nil {
		ks := cAlloc(len(keys) * int(ptrSize))
		list.keys = ks
		karr := unsafe.Slice((*unsafe.Pointer)(ks), len(keys))
		for i := range keys {
			karr[i] = cString(keys[i])
		}
	}

	return unsafe.Pointer(list)
}

func buildMap(m map[string]any) unsafe.Pointer {
	keys := make([]string, 0, len(m))
	values := make([]any, 0, len(m))
	for k, v := range m {
		keys = append(keys, k)
		values = append(values, v)
	}

	return buildList(values, keys)
}

// cString copies s into a NUL-terminated C heap buffer.
func cString(s string) unsafe.Pointer {
	b := cAlloc(len(s) + 1)
	dst := unsafe.Slice((*byte)(b), len(s)+1)
	copy(dst, s)
	dst[len(s)] = 0

	return b
}
