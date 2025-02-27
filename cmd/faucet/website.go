// Code generated by go-bindata.
// sources:
// faucet.html
// DO NOT EDIT!

package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _faucetHtml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xcc\x5a\x6d\x93\xdb\x36\x92\xfe\x3c\xfe\x15\x1d\x9e\xbd\x92\xce\x43\x52\x33\x63\x7b\x7d\x12\xa9\x94\xd7\x9b\xdd\xf3\xd5\x5d\x92\x4a\x9c\xba\xdb\xca\xa6\xae\x40\xb2\x25\xc2\x03\x02\x0c\x00\x4a\xa3\x4c\xe9\xbf\x5f\x35\x40\x52\xd4\xcb\x4c\xec\xb5\xaf\x6a\xfd\x61\x4c\x02\x8d\x46\xa3\xfb\x69\xf4\x0b\x95\x7c\xf5\xe7\xef\xde\xbe\xff\xdb\xf7\xdf\x40\x69\x2b\xb1\x78\x92\xd0\x7f\x20\x98\x5c\xa5\x01\xca\x60\xf1\xe4\x22\x29\x91\x15\x8b\x27\x17\x17\x49\x85\x96\x41\x5e\x32\x6d\xd0\xa6\x41\x63\x97\xe1\xeb\x60\x3f\x51\x5a\x5b\x87\xf8\x6b\xc3\xd7\x69\xf0\x3f\xe1\x4f\x6f\xc2\xb7\xaa\xaa\x99\xe5\x99\xc0\x00\x72\x25\x2d\x4a\x9b\x06\xef\xbe\x49\xb1\x58\xe1\x60\x9d\x64\x15\xa6\xc1\x9a\xe3\xa6\x56\xda\x0e\x48\x37\xbc\xb0\x65\x5a\xe0\x9a\xe7\x18\xba\x97\x4b\xe0\x92\x5b\xce\x44\x68\x72\x26\x30\xbd\x0a\x16\x4f\x88\x8f\xe5\x56\xe0\xe2\xfe\x3e\xfa\x16\xed\x46\xe9\xdb\xdd\x6e\x06\x6f\x1a\x5b\xa2\xb4\x3c\x67\x16\x0b\xf8\x0b\x6b\x72\xb4\x49\xec\x29\xdd\x22\xc1\xe5\x2d\x94\x1a\x97\x69\x40\xa2\x9b\x59\x1c\xe7\x85\xfc\x60\xa2\x5c\xa8\xa6\x58\x0a\xa6\x31\xca\x55\x15\xb3\x0f\xec\x2e\x16\x3c\x33\xb1\xdd\x70\x6b\x51\x87\x99\x52\xd6\x58\xcd\xea\xf8\x26\xba\x89\xfe\x18\xe7\xc6\xc4\xfd\x58\x54\x71\x19\xe5\xc6\x04\xa0\x51\xa4\x81\xb1\x5b\x81\xa6\x44\xb4\x01\xc4\x8b\x7f\x6c\xdf\xa5\x92\x36\x64\x1b\x34\xaa\xc2\xf8\x45\xf4\xc7\x68\xea\xb6\x1c\x0e\x3f\xbe\x2b\x6d\x6b\x72\xcd\x6b\x0b\x46\xe7\x1f\xbd\xef\x87\x5f\x1b\xd4\xdb\xf8\x26\xba\x8a\xae\xda\x17\xb7\xcf\x07\x13\x2c\x92\xd8\x33\x5c\x7c\x16\xef\x50\x2a\xbb\x8d\xaf\xa3\x17\xd1\x55\x5c\xb3\xfc\x96\xad\xb0\xe8\x76\xa2\xa9\xa8\x1b\xfc\x62\xfb\x3e\x64\xc3\x0f\xc7\x26\xfc\x12\x9b\x55\xaa\x42\x69\xa3\x0f\x26\xbe\x8e\xae\x5e\x47\xd3\x6e\xe0\x94\xbf\xdb\x80\x8c\x46\x5b\x5d\x44\x6b\xd4\x84\x5c\x11\xe6\x28\x2d\x6a\xb8\xa7\xd1\x8b\x8a\xcb\xb0\x44\xbe\x2a\xed\x0c\xae\xa6\xd3\x67\xf3\x73\xa3\xeb\xd2\x0f\x17\xdc\xd4\x82\x6d\x67\xb0\x14\x78\xe7\x87\x98\xe0\x2b\x19\x72\x8b\x95\x99\x81\xe7\xec\x26\x76\x6e\xcf\x5a\xab\x95\x46\x63\xda\xcd\x6a\x65\xb8\xe5\x4a\xce\x08\x51\xcc\xf2\x35\x9e\xa3\x35\x35\x93\x27\x0b\x58\x66\x94\x68\x2c\x1e\x09\x92\x09\x95\xdf\xfa\x31\xe7\xcd\xc3\x43\xe4\x4a\x28\x3d\x83\x4d\xc9\xdb\x65\xe0\x36\x82\x5a\x63\xcb\x1e\x6a\x56\x14\x5c\xae\x66\xf0\xaa\x6e\xcf\x03\x15\xd3\x2b\x2e\x67\x30\xdd\x2f\x49\xe2\x4e\x8d\x49\xec\x2f\xae\x27\x17\x49\xa6\x8a\xad\xb3\x61\xc1\xd7\x90\x0b\x66\x4c\x1a\x1c\xa9\xd8\x5d\x48\x07\x04\x74\x0f\x31\x2e\xbb\xa9\x83\x39\xad\x36\x01\xb8\x8d\xd2\xc0\x0b\x11\x66\xca\x5a\x55\xcd\xe0\x8a\xc4\x6b\x97\x1c\xf1\x13\xa1\x58\x85\x57\xd7\xdd\xe4\x45\x52\x5e\x75\x4c\x2c\xde\xd9\xd0\xd9\xa7\xb7\x4c\xb0\x48\x78\xb7\x76\xc9\x60\xc9\xc2\x8c\xd9\x32\x00\xa6\x39\x0b\x4b\x5e\x14\x28\xd3\xc0\xea\x06\x09\x47\x7c\x01\xc3\xeb\xef\x81\xdb\xaf\xbc\xea\xe4\x8a\x0b\xbe\x6e\x8f\x35\x78\x3c\x3a\xe1\xc3\x87\x78\x0d\xed\x83\x5a\x2e\x0d\xda\x70\x70\xa6\x01\x31\x97\x75\x63\xc3\x95\x56\x4d\xdd\xcf\x5f\x24\x6e\x14\x78\x91\x06\x8d\x16\x41\x7b\xfd\xbb\x47\xbb\xad\x5b\x55\x04\xfd\xc1\x95\xae\x42\xb2\x84\x56\x22\x80\x5a\xb0\x1c\x4b\x25\x0a\xd4\x69\xf0\xa3\xca\x39\x13\x20\xfd\x99\xe1\xa7\x1f\xfe\x13\x5a\x93\x71\xb9\x82\xad\x6a\x34\xbc\x67\xfa\xad\xe2\x12\x58\x51\x10\x5a\xa3\x28\x1a\xc8\xe1\xa0\x7b\x2a\x69\x98\x59\xb9\xa7\xba\x48\xb2\xc6\x5a\xd5\x13\x66\x56\x42\x66\x65\x58\xe0\x92\x35\xc2\x42\xa1\x55\x5d\xa8\x8d\x0c\xad\x5a\xad\x28\xd0\xf9\x33\xf8\x45\x01\x14\xcc\xb2\x76\x2a\x0d\x3a\xda\xce\x84\xcc\xd4\xaa\x6e\xea\xd6\x88\x7e\x10\xef\x6a\x26\x0b\x2c\xc8\xe4\xc2\x60\xb0\xf8\x2b\x5f\x23\x54\x08\xdf\xd8\x12\xf5\xc5\x31\x22\x72\xa6\xd1\x86\x43\xa6\x27\xb8\x48\x62\x2f\x8c\x3f\x12\xb4\xff\x92\x46\x74\x9c\xfa\x23\x54\x28\x1b\x38\x78\x0b\x35\x5d\x2b\xc1\xe2\xfe\x5e\x33\xb9\x42\x78\xca\x8b\xbb\x4b\x78\xca\x2a\xd5\x48\x0b\xb3\x14\xa2\x37\xee\xd1\xec\x76\x07\xdc\x01\x12\xc1\x17\x09\x7b\x0c\xdd\xa0\x64\x2e\x78\x7e\x9b\x06\x96\xa3\x4e\xef\xef\x89\xf9\x6e\x37\x87\xfb\x7b\xbe\x84\xa7\xd1\x0f\x98\xb3\xda\xe6\x25\xdb\xed\x56\xba\x7b\x8e\xf0\x0e\xf3\xc6\xe2\x78\x72\x7f\x8f\xc2\xe0\x6e\x67\x9a\xac\xe2\x76\xdc\x2d\xa7\x71\x59\xec\x76\x24\x73\x2b\xe7\x6e\x07\x31\x31\x95\x05\xde\xc1\xd3\xe8\x7b\xd4\x5c\x15\x06\x3c\x7d\x12\xb3\x45\x12\x0b\xbe\x68\xd7\x1d\x2a\x29\x6e\xc4\x1e\x2f\x31\x01\xa6\x87\xb9\xf3\x1a\x27\xea\x50\xd2\x33\x4e\xb0\x0a\x7b\xe9\x5b\x3c\x18\x6e\xf1\x16\xb7\x69\x70\x7f\x3f\x5c\xdb\xce\xe6\x4c\x88\x8c\x91\x5e\xfc\xd1\xfa\x45\xbf\x21\xe1\x74\xcd\x8d\xcb\xa8\x16\x9d\x04\x7b\xb1\x3f\xd2\xab\x8f\xee\x2d\xab\xea\x19\xdc\x5c\x0f\x2e\xad\x73\x0e\xff\xea\xc8\xe1\x6f\xce\x12\xd7\x4c\xa2\x00\xf7\x37\x34\x15\x13\xdd\x73\xeb\x2d\x03\xe7\x3b\x5e\x14\xd2\x15\xdd\x8b\xd6\x5f\xf5\xd3\x39\xa8\x35\xea\xa5\x50\x9b\x19\xb0\xc6\xaa\x39\x54\xec\xae\x0f\x77\x37\xd3\xe9\x50\x6e\xca\x04\x59\x26\xd0\x5d\x2e\x1a\x7f\x6d\xd0\x58\xd3\x5f\x25\x7e\xca\xfd\xa5\x1b\xa5\x40\x69\xb0\x38\xd2\x06\xed\x48\xaa\x75\x54\x03\xd3\xf7\xca\x3c\x2b\xfb\x52\xa9\x3e\x82\x0c\xc5\x68\x59\x0f\x82\x5d\xb0\x48\xac\xde\xd3\x5d\x24\xb6\xf8\xa4\x08\xa0\x29\xc3\x7b\x28\x00\xf8\x1b\x8d\xce\x5e\x23\x6a\x9f\x5e\x10\x64\xc1\xbd\x26\xb1\x2d\x3e\x63\x67\x02\x61\xc6\x0c\x7e\xcc\xf6\x2e\xd0\xef\xb7\x77\xaf\x9f\xbb\x7f\x89\x4c\xdb\x0c\x99\xfd\x18\x01\x96\x8d\x2c\x06\xe7\x77\x77\xe7\xe7\x0a\xd0\x48\xbe\x46\x6d\xb8\xdd\x7e\xac\x04\x58\xec\x45\xf0\xef\x87\x22\x24\xb1\xd5\x8f\x63\x6d\xf8\xf2\x85\x9c\xfb\xf7\x32\x92\x9b\xc5\xbf\xab\x0d\x14\x0a\x0d\xd8\x92\x1b\xa0\xd8\xfa\x75\x12\x97\x37\x3d\x49\xbd\x78\x4f\x13\x4e\xa9\xb0\x74\x99\x05\x70\x03\xba\x91\x2e\xf0\x2a\x09\xb6\xc4\xc3\x6c\xa4\x8d\xd1\x11\xbc\x57\x94\xd1\xad\x51\x5a\xa8\x98\xe0\x39\x57\x8d\x01\x96\x5b\xa5\x0d\x2c\xb5\xaa\x00\xef\x4a\xd6\x18\x4b\x8c\xe8\xfa\x60\x6b\xc6\x85\xf3\x25\x67\x52\x50\x1a\x58\x9e\x37\x55\x43\x19\xa9\x5c\x01\x4a\xd5\xac\xca\x56\x16\xab\xc0\x07\x26\xa1\xe4\xaa\x97\xc7\xd4\xac\x02\x66\x2d\xcb\x6f\xcd\x25\x74\xb7\x02\x30\x8d\x60\x39\x16\xb4\x2a\x57\x55\xa5\x24\xdc\xe8\x02\x6a\xa6\xed\x16\xcc\x61\x6a\xc1\xf2\xdc\x45\xb9\x08\xde\xc8\xad\x92\x08\x25\x5b\x3b\x09\xe1\xbd\xaf\x26\x48\xae\xbf\xb0\x1c\x33\xa5\x7a\x6a\xa8\xd8\xb6\xdb\xae\x95\x7e\xc3\x6d\xc9\xbd\x7a\x6a\xd4\x15\x2d\x2d\x40\xf0\x8a\x5b\x13\x25\x71\xbd\xbf\x51\xf7\xb1\x59\x84\xa5\xd2\xfc\x37\xca\x6b\xc4\xf0\xfa\xb4\x47\x97\x4b\x77\x37\x3a\xab\x0b\x5c\xda\x19\xbc\xf0\x77\xe3\x31\x8e\xdb\x02\xe8\x1c\x88\x3b\x9e\xae\xb0\xa4\x80\x33\x83\x1b\x9f\xcd\xfa\x44\xa2\xb0\x03\x09\x8a\x23\xa8\xf9\x4d\x5f\xbf\xae\xef\x7a\x39\xfa\x94\x78\xda\x33\x21\x04\x1c\x2a\x65\xcd\x7b\x35\x5e\x42\xc5\x6e\x11\x18\x24\xec\xa8\x40\x6e\x85\x76\xe5\x15\x77\xed\x81\xd8\x6e\x10\xed\xd7\xe4\xba\xe9\x0f\x9e\x21\x97\xab\x67\xd7\x53\x8f\x48\x7a\x20\xf6\xcf\xae\xa7\x5c\x5a\xf5\xec\x7a\x3a\xbd\x9b\x7e\xe4\xbf\x67\xd7\x53\x25\x9f\x5d\x4f\x6d\x89\xcf\xae\xa7\xcf\xae\x6f\x86\x58\xf6\x23\x6d\x62\x49\x44\x68\x68\xb3\x0e\xe1\x01\x58\xa6\x57\x68\xd3\xe0\x7f\x59\xa6\x1a\x3b\xcb\x04\x93\xb7\xc1\xc2\x49\x4b\xc9\x86\x03\xc1\xd9\xec\x14\x6a\x66\x08\x10\x24\xaf\xc3\x48\xdb\x08\x31\x30\x36\x8d\xd6\xaa\x91\x14\x13\x81\x4e\xec\xfc\x53\x8e\x08\x63\xa4\x96\x49\x94\x64\x3a\x5e\xbc\x55\xf5\x36\x74\x4c\xdc\xf2\x13\x25\x9a\xa6\xae\x95\xb6\xd1\x50\x99\x8c\x8a\x20\x81\x26\x7e\x3d\x7d\xf9\xfa\xd5\xa3\xd2\x1b\x4a\xb1\xdd\x11\x7a\x09\x59\xa6\xd6\x08\x3e\xa1\xcf\xd4\x1d\x30\x59\xc0\x92\x6b\x04\xb6\x61\xdb\xaf\x92\xb8\x70\xe5\xd7\xe7\x63\x76\xd9\xfa\xd6\x3f\x15\x68\x3b\x87\xbf\x84\xba\xc9\x04\x37\x25\x30\x90\xb8\x81\xc4\x58\xad\xe4\x6a\xe1\x46\x73\xaa\x47\xdd\x2b\xd4\xca\xd8\x47\xac\x8f\x55\x86\x45\x71\xc6\xfe\x5f\xca\xfc\x9b\xcd\x26\xea\x14\xe9\x6c\x5f\xa2\xa8\x63\xba\xfb\x1a\xc9\xed\x36\xf6\x3e\xa4\x64\xfc\x35\x2f\xd2\xeb\xd7\xd7\xaf\x5e\x5d\xbf\xf8\xb7\xd7\x2f\x5f\x5e\xbf\x7e\xf1\xf2\x21\x60\xd0\x99\x3e\x13\x17\x3e\x87\xfe\x56\x51\xc5\xda\x27\xd0\x1e\x2e\x5d\xe2\x46\xe1\xb9\xa0\x02\x44\x07\xff\x30\x84\x1a\x49\x59\x48\xc8\xc4\xd9\x04\xe2\x13\x40\xe4\x50\xf4\x88\x64\x9f\x89\xac\x0e\x3d\x04\x14\xd5\x58\x3a\x61\x57\xc8\x73\x25\x7b\x34\x5d\x82\xe1\x55\x2d\xb6\x90\xef\xad\x7e\x16\x56\x0f\xda\xe4\x77\x51\x75\x68\x35\x8f\x31\x17\xf9\x2b\x55\x20\x45\x7c\xd3\x98\x1c\x6b\xd7\xe0\xa5\x28\xfa\xa7\xed\x6f\x4c\x5a\x2e\xb1\x8b\xb6\x11\x7c\x27\xc5\x16\x1a\x83\xb0\x54\x1a\x0a\xcc\x9a\xd5\xca\xa5\x08\x1a\x6a\xcd\xd7\xcc\x62\x17\x62\x4d\x0b\x8a\x1e\x13\x83\xaa\x86\xd2\x1d\x31\xc8\x3e\xfe\xa6\x1a\xc8\x99\x04\xab\x59\x7e\xeb\x1d\xa5\xd1\x9a\x1c\xa5\x46\x7f\x9a\x3e\xc8\x67\x28\xd4\xc6\x91\xf8\x73\x2f\x39\x0a\x17\xf1\x0d\x22\x94\x6a\x03\x55\x93\x3b\x77\xa4\x88\xee\x0e\xb1\x61\xdc\x42\x23\x2d\x17\x5e\x9d\xb6\xd1\x92\xf2\x03\x3c\x88\xd0\x27\x75\x5f\x82\xd5\xe2\x7d\x89\x67\xd2\xa1\xbe\x62\x03\x8d\x6f\x3d\x39\xd4\x5a\x59\xcc\xc9\x9e\xc0\x56\x8c\x4b\x43\x16\x71\x39\x00\x56\x1f\x51\xd1\xf5\x4f\xed\xc3\xbe\x39\xe9\xa6\xe3\x18\xfe\x2a\x54\xc6\x04\xac\x09\xe8\x99\xa0\x54\x4e\x41\xa9\xe8\xe8\x03\x6d\x19\xcb\x6c\x63\x40\x2d\xdd\xa8\x97\x9c\xd6\xaf\x99\x26\x0b\x62\x55\x5b\x48\xdb\xd6\x1a\x8d\x19\xd4\xeb\xb6\x61\x48\xaf\x54\xb5\x1f\xcc\xf7\x5a\x4f\xe1\xe7\x5f\xe6\x4f\x5a\x51\xfe\x8c\x4b\x07\x09\x82\xb7\x3f\xb2\x2d\x99\x85\x5c\x23\xb3\x68\x20\x17\xca\x34\xda\x4b\x58\x68\x55\x03\x49\xd9\x71\xea\x38\xd3\x44\xed\x76\xeb\x98\x8c\x4b\x66\xca\x49\xdb\x19\xd4\xe8\xac\xd4\xcf\x75\xe3\x17\x84\xba\x31\x31\xe0\xe9\x74\x0e\x3c\xe9\xf8\x46\x02\xe5\xca\x96\x73\xe0\xcf\x9f\xf7\xc4\x17\x7c\x09\xe3\x8e\xe2\x67\xfe\x4b\x64\xef\x22\xda\x05\xd2\x14\x86\xbb\xb9\x0d\x5b\x3e\xa6\x16\x3c\xc7\x31\xbf\x84\xab\xc9\xbc\x9b\xcd\x34\xb2\xdb\xee\xad\xb5\xa3\xff\xcf\xfd\xdd\xcd\x0f\x35\xe3\x94\x7f\xa0\x1b\x5f\xf7\x1b\x60\xb0\xe2\xc6\x42\xa3\x05\xb4\x3e\xec\x4d\xd0\x1b\xc4\xd1\x0d\xb5\x72\x82\xcb\xf6\xa1\xc5\x54\x77\x04\xcf\x26\x32\x28\x8b\xf1\x7f\xfc\xf8\xdd\xb7\x91\xb1\x9a\xcb\x15\x5f\x6e\xc7\xf7\x8d\x16\x33\x78\x3a\x0e\xfe\xa5\xd1\x22\x98\xfc\x3c\xfd\x25\x5a\x33\xd1\xe0\xa5\xb3\xf7\xcc\xfd\x3d\xd9\xe5\x12\xda\xc7\x19\x1c\x6e\xb8\x9b\x4c\xe6\xe7\x7b\x24\x83\x96\x8e\x46\x83\x76\x4c\x84\x3d\xf0\x8f\x75\xc4\xa0\x42\x5b\x2a\xe7\xba\x1a\x73\x25\x25\xe6\x16\x9a\x5a\xc9\x56\x25\x20\x94\x31\x7b\x20\x76\x14\xe9\x29\x28\x5a\xfa\xd4\x85\xea\xff\xc6\xec\x47\x95\xdf\xa2\x1d\x8f\xc7\x1b\x2e\x0b\xb5\x89\x84\xf2\x37\x6d\x44\x4e\xaa\x72\x25\x20\x4d\x53\x68\x83\x68\x30\x81\xaf\x21\xd8\x18\x0a\xa7\x01\xcc\xe8\x91\x9e\x26\xf0\x1c\x8e\x97\x97\x14\xed\x9f\x43\x10\xb3\x9a\x07\x13\xef\x0e\x9d\xe2\x95\xac\xd0\x18\xb6\xc2\xa1\x80\xae\x2a\xea\x41\x46\xe7\xa8\xcc\x0a\x52\x70\x06\xaa\x99\x36\xe8\x49\x22\xaa\xc4\x3b\xb4\x11\x66\x1d\x59\x9a\x82\x6c\x84\xd8\x83\xd4\x3b\xc5\xbc\x83\xdf\x01\x79\xe4\x43\xcd\x57\x69\x0a\x54\x96\x92\x8a\x8b\xfd\x4a\x32\xbe\x2f\xa0\x27\x11\xc5\x85\xfd\x8a\xc9\x7c\x88\xe6\x03\x6e\x58\xfc\x1e\x3b\x2c\x8e\xf9\x61\xf1\x00\x43\xd7\xaf\x78\x8c\x9f\xef\x6f\x0c\xd8\xb9\x81\x07\xb8\xc9\xa6\xca\x50\x3f\xc6\xce\xf7\x2b\x5a\x76\x4e\xd5\xef\xa4\x1d\xac\xbd\x84\xab\x57\x93\x07\xb8\xa3\xd6\xea\x41\xe6\x52\xd9\xed\xf8\x5e\xb0\x2d\xa5\x4c\x30\xb2\xaa\x7e\xeb\xda\x0b\xa3\x4b\x17\x71\x67\xd0\x73\xb8\x74\x8d\xe3\x19\x8c\xdc\x1b\xcd\xf3\x0a\xdd\xaa\x97\xd3\xe9\xf4\x12\xba\x0f\x2e\x7f\x62\xe4\x84\xba\xc1\xdd\x03\xf2\x98\x26\xcf\x29\xee\x7f\x8e\x44\x2d\x8f\x5e\xa6\xf6\xfd\x33\xa4\xea\x63\xc3\x81\x58\xf0\x87\x3f\xc0\xc9\xec\x21\x8c\xe3\x18\xfe\x8b\x51\x09\x2e\x84\xeb\x1c\xb8\x86\x41\x4f\x5f\x71\x63\x5c\x21\x6e\xa0\x50\x12\xdb\x35\x9f\x76\xed\x9f\xc8\xd8\x92\xc1\x02\xa6\xc7\x02\xd2\x75\x38\x08\x0b\x67\xa2\xc5\x80\xef\x61\x20\xb8\xd8\x0d\xf7\x3b\x58\xc9\x2b\x84\xaf\x52\x08\x82\xe1\xe2\x13\x0a\x22\xe8\x99\x5d\x18\xb4\xef\xbd\x2d\xc6\x6d\x74\x3c\x17\xbb\x26\x97\x70\x33\x9d\x4e\x27\x27\x42\xec\xf6\xea\x7d\x53\x53\xda\x04\x4c\x6e\xdd\x95\xd8\xeb\xd6\x25\x8e\x94\x02\xd1\x95\x26\x20\x57\x42\xf8\x9c\xa5\x5d\x4a\x0a\x6e\x1b\x27\x29\x84\x57\xf3\x33\x51\x74\xa0\xc9\xc1\xd1\x8e\xcd\x73\x46\xf7\xc7\x26\x3a\xd4\xd9\x11\x71\x78\x75\x60\x94\x03\x7b\x9d\x37\xcc\x45\x2f\x37\xdf\x6b\xf4\xc8\x5c\x7b\x7b\x1d\xeb\x6c\x20\xbf\xe7\xf3\xfc\xea\x23\x8f\xd1\x4f\xd7\x8d\x29\xc7\x47\x82\x4e\xe6\xa7\xb6\x79\x67\x51\x53\x96\xac\x28\x64\x91\x2d\xa8\x12\xd0\x78\x62\x12\x97\xaa\x6b\x0c\x35\xca\x02\x75\x97\x52\xf8\xcc\x9e\x12\xc0\x03\x93\xf9\xa2\x72\x08\xa7\x4f\x74\x18\x97\x92\x29\x89\x00\x00\x47\x4e\xe0\x80\x7a\x80\x54\x22\x46\xc1\x6a\x83\x05\xa4\xe0\xbf\x7f\x8f\x27\x51\x23\xf9\xdd\x78\x12\xb6\xef\xc7\x3c\xba\xf9\x79\x5f\x25\x76\x62\x3f\x4f\x21\x48\xac\x06\x5e\xa4\xa3\x00\x9e\x9f\x73\x41\x8a\xba\xa3\xc5\x5e\x82\xe1\x52\x80\xc4\x16\x0b\xd7\x03\xf5\xe5\xda\xdf\x83\x8c\xe5\xb7\x2b\x57\x08\xcd\x28\xd5\x1a\x9f\xb0\x65\x6b\x66\x99\x76\x5c\x27\x73\xd8\x93\xb7\x75\x62\x4e\xc6\x99\x83\x2f\x48\x5d\xab\x15\xfa\xcf\x13\xee\x2d\x53\xba\x40\x1d\x6a\x56\xf0\xc6\xcc\xe0\x45\x7d\x37\xff\x7b\xf7\xf9\xc6\x35\x84\x1f\x15\xb5\xd6\xb8\x38\x91\xa8\xed\x30\x3e\x87\x20\x89\x89\xe0\xf7\xd8\xf4\x87\x1d\x7e\x77\x87\x33\x6d\x6f\xe8\xbf\x8a\xb7\xe3\x15\x2f\x0a\x81\x24\xf0\x9e\x3d\x39\x23\xd9\x7f\xe8\x52\x87\x5b\x42\xdb\xef\xde\xaf\xd9\x01\x0a\x83\x8f\x2c\xe8\x5b\xe7\x23\x02\x40\x48\x47\xe6\x4e\xe7\x6d\xad\xed\x86\xf5\xc8\xe9\xa2\xfd\x15\x45\xd1\x68\x97\x6b\x8d\xc3\x16\x60\x97\x30\x32\x94\xfb\x15\x66\x34\x89\xca\xa6\x62\x92\xff\x86\x63\x8a\x4b\x13\xaf\x2b\xd7\x8b\x0f\x4e\xaf\xe4\x13\x61\xf6\x4d\xf2\x51\x17\xe3\x46\xad\x12\x47\x9d\x75\x5f\xec\x4b\xfb\x19\x4c\xe7\xa3\x4f\xd4\xd0\xf9\x5d\xc2\x8c\x69\x18\xbe\x84\x5d\xf0\x05\xad\x68\xf7\x6e\x2e\x63\x7a\xe4\x1b\x19\x2e\x3f\x97\x6a\x93\x8e\x6e\xa6\xbd\x90\xde\xd0\xce\xce\xa3\x16\x6b\x27\xc6\x20\x29\x3b\xd7\x5c\xc0\xcd\xf4\x4b\x48\xeb\x9b\x21\x47\x27\xb0\x9a\xd7\x58\x00\xcb\x2d\x5f\xe3\xff\xc3\x41\xbe\x80\x92\x3f\x59\x44\xc2\x61\xa7\x3c\x07\xd3\x03\x79\x69\xb6\xd7\xed\xbf\x92\xbf\x41\xec\x34\xfc\x1c\x82\xb3\x07\x79\x10\x89\x47\x84\x47\xae\xfd\xb0\xdf\xbb\x8f\x4b\xc1\x71\x4c\xa1\x6c\xb7\xff\x30\x3a\x89\x4a\x5b\x89\x71\x90\x58\xf7\xfb\x18\x92\xb9\xe7\xe0\x18\xf8\xe1\xc3\x94\x6e\x77\x58\xc8\x50\xfd\x8e\x47\x75\x16\x0c\x92\x93\xbe\x16\xeb\x32\x11\xd8\xed\x7f\x46\x14\xc7\xf0\xa3\x65\xda\x02\x83\x9f\xde\x41\x53\x17\xcc\xfa\xcf\x38\x14\x1f\xfd\x67\x92\xee\x77\x46\x19\xd3\x06\x96\x4a\x6f\x98\x2e\xda\xfe\x8c\x2d\x71\xeb\x3e\xe3\x74\xa9\x9f\x41\xfb\x8e\x6e\xb1\x35\x13\xe3\x93\xba\xef\xe9\x78\x14\x0d\x4d\x3e\x9a\x44\xc8\xf2\xf2\x94\xd0\x45\xac\x7e\xdf\x14\xbe\x75\x25\xc0\xf8\xe9\xd8\x96\xdc\x4c\x22\x66\xad\x1e\x8f\x0e\xc0\x30\x9a\x90\x5d\xaf\x06\x25\x59\xbf\x3c\x39\x70\xab\xc7\x78\xec\x93\xe9\x3e\x11\xe8\xc8\x73\x63\xc6\x1e\x57\xa3\xcb\x01\xef\x43\x58\x8d\x9e\x8d\x7a\x43\xed\xdd\x7b\x7f\x8e\xf4\xac\x24\x07\xac\x47\xe4\x65\xa3\x93\xed\x59\x51\xbc\x25\xff\x19\x07\x67\x3c\xfd\x18\x1d\x93\x5e\xd9\xfe\xbe\x7e\x54\xcb\xfe\x27\x19\x0f\xa8\x98\x17\xa3\x49\x64\x9a\xcc\xf7\x26\xc6\x2f\xfb\x02\xac\x23\x73\xe0\x3d\x0e\x05\x27\x09\x05\x6d\x71\x98\x54\x84\x47\x49\xc8\x23\x51\xa3\xdd\xd2\x9f\x6a\x77\x49\x0a\x9f\x4e\xfa\xd6\xd6\x37\x86\x92\x2b\xdf\xf8\xdf\x60\x66\x5c\x27\x01\x5a\xbc\xbb\x6e\x8e\xef\xda\xbc\xf9\xfe\xdd\xa0\x73\xd3\x7b\xc4\xd8\x71\xef\x7f\x02\x78\xae\x4f\x72\xf6\x37\x87\x9b\xcd\x26\x5a\x29\xb5\x12\xfe\xd7\x86\x7d\x23\x25\x66\x35\x8f\x3e\x98\x00\x98\xd9\xca\x1c\x0a\x5c\xa2\x5e\x0c\xd8\xb7\xdd\x95\x24\xf6\xbf\x86\x4b\x62\xff\x83\xdf\xff\x0b\x00\x00\xff\xff\x08\x83\x35\x21\x01\x2c\x00\x00")

func faucetHtmlBytes() ([]byte, error) {
	return bindataRead(
		_faucetHtml,
		"faucet.html",
	)
}

func faucetHtml() (*asset, error) {
	bytes, err := faucetHtmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "faucet.html", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
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
var _bindata = map[string]func() (*asset, error){
	"faucet.html": faucetHtml,
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
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"faucet.html": {faucetHtml, map[string]*bintree{}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
