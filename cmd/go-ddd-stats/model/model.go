package model

import "fmt"

type Stats struct {
	Files     []*File
	Packages  []*Package
	Histogram []*Size
}

type File struct {
	Name    string
	Path    string
	Package string
	Size    int64
}

type Package struct {
	Name    string
	Path    string
	Size    int64
	Count   int64
	Average int64
}

type Size struct {
	Size  string
	Count int64
}

func (m *Stats) AppendFile(f *File) {
	m.Files = append(m.Files, f)
}

func (m *Stats) Package(packagePath string) *Package {
	for _, pkg := range m.Packages {
		if pkg.Path == packagePath {
			return pkg
		}
	}
	result := &Package{
		Path: packagePath,
	}
	m.Packages = append(m.Packages, result)
	return result
}

func Histogram(m *Stats) []*Size {
	bucketLimits := []int64{
		1 << 10,   // 1 KB
		2 << 10,   // 2 KB
		4 << 10,   // 4 KB
		8 << 10,   // 8 KB
		16 << 10,  // 16 KB
		32 << 10,  // 32 KB
		64 << 10,  // 64 KB
		128 << 10, // 128 KB
		256 << 10, // 256 KB
	}

	// Prepare bucket labels and counters
	buckets := make([]*Size, len(bucketLimits)+1)
	for i := 0; i < len(bucketLimits); i++ {
		buckets[i] = &Size{
			Size:  fmt.Sprintf("< %d KB", bucketLimits[i]>>10),
			Count: 0,
		}
	}
	// Last bucket is for sizes > 256KB
	buckets[len(bucketLimits)] = &Size{
		Size:  "> 256 KB",
		Count: 0,
	}

	for _, file := range m.Files {
		sz := file.Size
		placed := false
		for i, limit := range bucketLimits {
			if sz < limit {
				buckets[i].Count++
				placed = true
				break
			}
		}
		if !placed {
			// Size > 256 KB
			buckets[len(bucketLimits)].Count++
		}
	}

	return buckets
}
