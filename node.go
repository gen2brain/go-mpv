package mpv

/*
#include <mpv/client.h>
#include <stdlib.h>

int GetNodeFlag(mpv_node* node) {
    return (int)node->u.flag;
}

char* GetNodeString(mpv_node* node) {
    return (char*)node->u.string;
}

int64_t GetNodeInt64(mpv_node* node) {
    return (int64_t)node->u.int64;
}

double GetNodeDouble(mpv_node* node) {
    return (double)node->u.double_;
}

mpv_node_list* GetNodeList(mpv_node* node) {
    return (mpv_node_list*)node->u.list;
}

mpv_byte_array* GetNodeByteArray(mpv_node* node) {
    return (mpv_byte_array*)node->u.ba;
}

mpv_node* GetNodeListIndex(mpv_node* node_list, int i) {
    return &node_list[i];
}

char* GetCharArrayIndex(char** arr, int i) {
    return arr[i];
}

void SetNodeFlag(mpv_node* node, int i) {
    node->u.flag = i;
}

void SetNodeString(mpv_node* node, char* str) {
    node->u.string = str;
}

void SetNodeInt64(mpv_node* node, int64_t i) {
    node->u.int64 = i;
}

void SetNodeDouble(mpv_node* node, double d) {
    node->u.double_ = d;
}

void SetNodeList(mpv_node* node, mpv_node_list* list) {
    node->u.list = list;
}

void SetNodeByteArray(mpv_node* node, mpv_byte_array* ba) {
    node->u.ba = ba;
}

mpv_node_list* makeNodeList(int size) {
    mpv_node_list* list = malloc(sizeof(mpv_node_list));
    if (list == NULL) {
        return NULL;
    }
    list->values = calloc(sizeof(mpv_node), size);
    if (list->values == NULL) {
        return NULL;
    }
    list->keys = calloc(sizeof(char*), size);
    if (list->keys == NULL) {
        return NULL;
    }
    return list;
}

void setNodeListIndex(mpv_node_list* list, int i, mpv_node* node) {
    list->values[i] = *node;
}

void setStringArray(char** a, int i, char* s);
char** makeCharArray(int size);

*/
import "C"

import (
	"unsafe"
)

// Node type.
type Node struct {
	Data   interface{}
	Format Format
}

// NewNode takes a pointer to an mpv_node struct and returns a pointer to a native
// Go Node struct. The data is converted into Go data according to mpv_node.format.
func NewNode(n *C.mpv_node) *Node {
	switch Format(n.format) {
	case FORMAT_NONE:
		return &Node{nil, FORMAT_NONE}
	case FORMAT_STRING:
		return &Node{C.GoString(C.GetNodeString(n)), FORMAT_STRING}
	case FORMAT_FLAG:
		return &Node{(int(C.GetNodeFlag(n)) == 1), FORMAT_FLAG}
	case FORMAT_INT64:
		return &Node{int64(C.GetNodeInt64(n)), FORMAT_INT64}
	case FORMAT_DOUBLE:
		return &Node{float64(C.GetNodeDouble(n)), FORMAT_DOUBLE}
	case FORMAT_NODE_MAP:
		return &Node{NewNodeMap(C.GetNodeList(n)), FORMAT_NODE_MAP}
	case FORMAT_NODE_ARRAY:
		return &Node{NewNodeList(C.GetNodeList(n)), FORMAT_NODE_ARRAY}
	case FORMAT_BYTE_ARRAY:
		ba := C.GetNodeByteArray(n)
		return &Node{C.GoBytes(unsafe.Pointer(ba), C.int(ba.size)), FORMAT_BYTE_ARRAY}
	default:
		return nil
	}
}

// CNode turns a Go Node into an mpv_node struct.
func (n *Node) CNode() *C.mpv_node {
	result := &C.mpv_node{}
	result.format = C.mpv_format(n.Format)
	switch n.Format {
	case FORMAT_NONE:
		return result
	case FORMAT_STRING:
		C.SetNodeString(result, C.CString(n.Data.(string)))
	case FORMAT_FLAG:
		if n.Data.(bool) {
			C.SetNodeFlag(result, C.int(1))
		} else {
			C.SetNodeFlag(result, C.int(0))
		}
	case FORMAT_INT64:
		C.SetNodeInt64(result, C.int64_t(n.Data.(int64)))
	case FORMAT_DOUBLE:
		C.SetNodeDouble(result, C.double(n.Data.(float64)))
	case FORMAT_NODE_MAP:
		C.SetNodeList(result, CNodeMap(n.Data.(map[string]*Node)))
	case FORMAT_NODE_ARRAY:
		C.SetNodeList(result, CNodeList(n.Data.([]*Node)))
	case FORMAT_BYTE_ARRAY:
		d := n.Data.([]byte)
		ba := &C.mpv_byte_array{}
		ba.size = C.size_t(len(d))
		ba.data = unsafe.Pointer(&d)
		C.SetNodeByteArray(result, ba)
	default:
		return nil
	}
	return result
}

// NewNodeList turns an mpv_node_list into a list of Go Nodes.
func NewNodeList(n *C.mpv_node_list) []*Node {
	nodes := make([]*Node, int(n.num))
	for i := 0; i < int(n.num); i++ {
		nodes[i] = NewNode(C.GetNodeListIndex(n.values, C.int(i)))
	}
	return nodes
}

// CNodeList turns a list of Go Nodes into an mpv_node_list
func CNodeList(nodelist []*Node) *C.mpv_node_list {
	result := C.makeNodeList(C.int(len(nodelist) + 1))
	if result == nil {
		return nil
	}
	for i, n := range nodelist {
		C.setNodeListIndex(result, C.int(i), n.CNode())
	}
	result.num = C.int(len(nodelist))
	return result
}

// NewNodeMap turns an mpv_node_list into a map of strings to Go Nodes.
func NewNodeMap(n *C.mpv_node_list) map[string]*Node {
	nodes := NewNodeList(n)
	nodemap := make(map[string]*Node)
	for i, node := range nodes {
		nodemap[C.GoString(C.GetCharArrayIndex(n.keys, C.int(i)))] = node
	}
	return nodemap
}

// CNodeMap turns a map of strings to Go nodes into an mpv_node_list.
func CNodeMap(nodemap map[string]*Node) *C.mpv_node_list {
	nodelist := make([]*Node, len(nodemap))
	keys := make([]string, len(nodemap))
	i := 0
	for k, v := range nodemap {
		nodelist[i] = v
		keys[i] = k
		i++
	}
	result := CNodeList(nodelist)
	if result == nil {
		return nil
	}
	result.keys = C.makeCharArray(C.int(len(keys)))
	for i, s := range keys {
		C.setStringArray(result.keys, C.int(i), C.CString(s))
	}
	return result
}
