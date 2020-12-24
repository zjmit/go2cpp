// SPDX-License-Identifier: Apache-2.0

package gowasm2cpp

import (
	"os"
	"path/filepath"
	"text/template"
)

type wasmData struct {
	Offset int
	Data   []byte
}

func writeMem(dir string, incpath string, namespace string, initPageNum int, data []wasmData) error {
	{
		f, err := os.Create(filepath.Join(dir, "mem.h"))
		if err != nil {
			return err
		}
		defer f.Close()

		if err := memHTmpl.Execute(f, struct {
			IncludeGuard string
			IncludePath  string
			Namespace    string
		}{
			IncludeGuard: includeGuard(namespace) + "_MEM_H",
			IncludePath:  incpath,
			Namespace:    namespace,
		}); err != nil {
			return err
		}
	}
	{
		f, err := os.Create(filepath.Join(dir, "mem.cpp"))
		if err != nil {
			return err
		}
		defer f.Close()

		if err := memCppTmpl.Execute(f, struct {
			IncludePath string
			Namespace   string
			InitPageNum int
			Data        []wasmData
		}{
			IncludePath: incpath,
			Namespace:   namespace,
			InitPageNum: initPageNum,
			Data:        data,
		}); err != nil {
			return err
		}
	}
	return nil
}

var memHTmpl = template.Must(template.New("mem.h").Parse(`// Code generated by go2cpp. DO NOT EDIT.

#ifndef {{.IncludeGuard}}
#define {{.IncludeGuard}}

#include <cstdint>
#include <string>
#include <vector>
#include "{{.IncludePath}}bytes.h"

namespace {{.Namespace}} {

class Mem {
public:
  static constexpr int32_t kPageSize = 64 * 1024;

  Mem();

  int32_t GetSize() const;
  int32_t Grow(int32_t delta);

  inline __attribute__((always_inline)) int8_t LoadInt8(int32_t addr) const {
    return static_cast<int8_t>(*(bytes_begin_ + addr));
  }

  inline __attribute__((always_inline)) uint8_t LoadUint8(int32_t addr) const {
    return *(bytes_begin_ + addr);
  }

  inline __attribute__((always_inline)) int16_t LoadInt16(int32_t addr) const {
    return *(reinterpret_cast<const int16_t*>(bytes_begin_ + addr));
  }

  inline __attribute__((always_inline)) uint16_t LoadUint16(int32_t addr) const {
    return *(reinterpret_cast<const uint16_t*>(bytes_begin_ + addr));
  }

  inline __attribute__((always_inline)) int32_t LoadInt32(int32_t addr) const {
    return *(reinterpret_cast<const int32_t*>(bytes_begin_ + addr));
  }

  inline __attribute__((always_inline)) uint32_t LoadUint32(int32_t addr) const {
    return *(reinterpret_cast<const uint32_t*>(bytes_begin_ + addr));
  }

  inline __attribute__((always_inline)) int64_t LoadInt64(int32_t addr) const {
    return *(reinterpret_cast<const int64_t*>(bytes_begin_ + addr));
  }

  inline __attribute__((always_inline)) float LoadFloat32(int32_t addr) const {
    return *(reinterpret_cast<const float*>(bytes_begin_ + addr));
  }

  inline __attribute__((always_inline)) double LoadFloat64(int32_t addr) const {
    return *(reinterpret_cast<const double*>(bytes_begin_ + addr));
  }

  void StoreInt8(int32_t addr, int8_t val) {
    *(bytes_begin_ + addr) = static_cast<uint8_t>(val);
  }

  inline __attribute__((always_inline)) void StoreInt16(int32_t addr, int16_t val) {
    *(reinterpret_cast<int16_t*>(bytes_begin_ + addr)) = val;
  }

  inline __attribute__((always_inline)) void StoreInt32(int32_t addr, int32_t val) {
    *(reinterpret_cast<int32_t*>(bytes_begin_ + addr)) = val;
  }

  inline __attribute__((always_inline)) void StoreInt64(int32_t addr, int64_t val) {
    *(reinterpret_cast<int64_t*>(bytes_begin_ + addr)) = val;
  }

  inline __attribute__((always_inline)) void StoreFloat32(int32_t addr, float val) {
    *(reinterpret_cast<float*>(bytes_begin_ + addr)) = val;
  }

  inline __attribute__((always_inline)) void StoreFloat64(int32_t addr, double val) {
    *(reinterpret_cast<double*>(bytes_begin_ + addr)) = val;
  }

  void StoreBytes(int32_t addr, const std::vector<uint8_t>& bytes);

  BytesSpan LoadSlice(int32_t addr);
  BytesSpan LoadSliceDirectly(int64_t array, int32_t len);
  std::string LoadString(int32_t addr) const;

private:
  Mem(const Mem&) = delete;
  Mem& operator=(const Mem&) = delete;

  std::vector<uint8_t> bytes_;
  uint8_t* bytes_begin_;
};

}

#endif  // {{.IncludeGuard}}
`))

var memCppTmpl = template.Must(template.New("mem.cpp").Parse(`// Code generated by go2cpp. DO NOT EDIT.

#include "{{.IncludePath}}mem.h"

#include <algorithm>
#include <cstring>

namespace {{.Namespace}} {

namespace {

{{range $index, $value := .Data}}const uint8_t data_segment_data{{$index}}[] = {
  {{range $value2 := $value.Data}}{{$value2}}, {{end}}
};
{{end}}
}

Mem::Mem() {
  // Reserving 4GB memory might fail on some consoles. 1GB should be safe in most environments.
  bytes_.reserve(1ul * 1024 * 1024 * 1024);
  bytes_.resize({{.InitPageNum}} * kPageSize);
  bytes_begin_ = &*bytes_.begin();
{{range $index, $value := .Data}}  std::memcpy(&(bytes_[{{$value.Offset}}]), data_segment_data{{$index}}, {{len $value.Data}});
{{end}}
}

int32_t Mem::GetSize() const {
  return bytes_.size() / kPageSize;
}

int32_t Mem::Grow(int32_t delta) {
  constexpr size_t kMaxMemorySizeOnWasm = 4ul * 1024ul * 1024ul * 1024ul;

  int prev_size = GetSize();
  size_t new_size = (prev_size + delta) * kPageSize;
  if (bytes_.capacity() < new_size) {
    size_t new_capacity = bytes_.capacity();
    while (new_capacity < new_size) {
      new_capacity *= 2;
    }
    new_capacity = std::min(new_capacity, kMaxMemorySizeOnWasm);
    bytes_.reserve(new_capacity);
    bytes_begin_ = &*bytes_.begin();
  }
  bytes_.resize(new_size);
  return prev_size;
}

void Mem::StoreBytes(int32_t addr, const std::vector<uint8_t>& bytes) {
  std::copy(bytes.begin(), bytes.end(), bytes_.begin() + addr);
}

BytesSpan Mem::LoadSlice(int32_t addr) {
  int64_t array = LoadInt64(addr);
  int64_t len = LoadInt64(addr + 8);
  return BytesSpan{&*(bytes_.begin() + array), static_cast<BytesSpan::size_type>(len)};
}

BytesSpan Mem::LoadSliceDirectly(int64_t array, int32_t len) {
  return BytesSpan{&*(bytes_.begin() + array), static_cast<BytesSpan::size_type>(len)};
}

std::string Mem::LoadString(int32_t addr) const {
  int64_t saddr = LoadInt64(addr);
  int64_t len = LoadInt64(addr + 8);
  return std::string{bytes_.begin() + saddr, bytes_.begin() + saddr + len};
}

}
`))
