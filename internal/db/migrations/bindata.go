package migrations

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strings"
)

func bindata_read(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	return buf.Bytes(), nil
}

var __000001_create_metricss_table_down_sql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\x72\x75\xf7\xf4\xb3\xe6\xe2\x72\x09\xf2\x0f\x50\x08\x71\x74\xf2\x71\x55\xf0\x74\x53\x70\x8d\xf0\x0c\x0e\x09\x56\xc8\x4d\x2d\x29\xca\x4c\x2e\xb6\x86\xca\x46\x06\xb8\x2a\xa4\xe6\x95\xe6\xc6\x97\x54\x16\xa4\x16\x5b\x73\x71\x39\xfb\xfb\xfa\x7a\x86\x58\x03\x02\x00\x00\xff\xff\x37\x77\xef\xac\x44\x00\x00\x00")

func _000001_create_metricss_table_down_sql() ([]byte, error) {
	return bindata_read(
		__000001_create_metricss_table_down_sql,
		"000001_create_metricss_table.down.sql",
	)
}

var __000001_create_metricss_table_up_sql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x54\xce\x31\x6b\xc3\x30\x10\x05\xe0\x5d\xbf\xe2\x6d\x4e\x20\x43\x97\x4e\x99\x64\xf7\x9a\x1e\xad\x24\x23\x29\xa5\x9e\x82\x49\x84\x31\xd8\x6e\xb1\xa5\x42\xff\x7d\x31\x5a\xec\xf1\xf1\xee\x1e\x5f\x49\x17\xd6\x67\x21\x2a\x4b\xd2\x13\x7c\x53\x13\xc2\x94\xc6\x5b\xfc\xfb\x09\x0b\xa4\x03\xe9\xab\xc2\x41\x00\x40\x71\xff\x4e\x53\x0c\x73\x71\xca\xb1\x6b\x53\x17\x0a\x71\xdc\xfc\xcb\xf2\x83\xc0\xaf\xd0\xc6\x83\xbe\xd8\x79\x87\x31\xc4\xb9\xbf\x2f\x79\x22\x87\x5b\xff\xc0\xa7\xb4\xd5\x9b\xb4\x38\x3c\x3f\x1d\x51\x5b\x56\xd2\x36\x78\xa7\xe6\xb4\xbd\x5b\x15\x1b\xcf\xae\x7b\x84\x21\xb6\x28\xf9\xc2\xda\xef\x8a\xdf\x76\x48\x01\x2f\xe6\xba\x5a\x6a\x4b\x15\x3b\x36\x3a\x33\x8d\x52\xec\xcf\xff\x01\x00\x00\xff\xff\xe4\x35\x8d\xf2\xf6\x00\x00\x00")

func _000001_create_metricss_table_up_sql() ([]byte, error) {
	return bindata_read(
		__000001_create_metricss_table_up_sql,
		"000001_create_metricss_table.up.sql",
	)
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		return f()
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() ([]byte, error){
	"000001_create_metricss_table.down.sql": _000001_create_metricss_table_down_sql,
	"000001_create_metricss_table.up.sql":   _000001_create_metricss_table_up_sql,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for name := range node.Children {
		rv = append(rv, name)
	}
	return rv, nil
}

type _bintree_t struct {
	Func     func() ([]byte, error)
	Children map[string]*_bintree_t
}

var _bintree = &_bintree_t{nil, map[string]*_bintree_t{
	"000001_create_metricss_table.down.sql": &_bintree_t{_000001_create_metricss_table_down_sql, map[string]*_bintree_t{}},
	"000001_create_metricss_table.up.sql":   &_bintree_t{_000001_create_metricss_table_up_sql, map[string]*_bintree_t{}},
}}
