// SPDX-License-Identifier: Apache-2.0

package gowasm2cpp

import (
	"os"
	"path/filepath"
	"text/template"
)

func writeJS(dir string, namespace string) error {
	{
		f, err := os.Create(filepath.Join(dir, "js.h"))
		if err != nil {
			return err
		}
		defer f.Close()

		if err := jsHTmpl.Execute(f, struct {
			IncludeGuard string
			Namespace    string
		}{
			IncludeGuard: includeGuard(namespace) + "_JS_H",
			Namespace:    namespace,
		}); err != nil {
			return err
		}
	}
	{
		f, err := os.Create(filepath.Join(dir, "js.cpp"))
		if err != nil {
			return err
		}
		defer f.Close()

		if err := jsCppTmpl.Execute(f, struct {
			Namespace string
		}{
			Namespace: namespace,
		}); err != nil {
			return err
		}
	}
	return nil
}

var jsHTmpl = template.Must(template.New("js.h").Parse(`// Code generated by go2cpp. DO NOT EDIT.

#ifndef {{.IncludeGuard}}
#define {{.IncludeGuard}}

#include <deque>
#include <functional>
#include <iostream>
#include <map>
#include <memory>
#include <string>
#include <vector>

namespace {{.Namespace}} {

class Writer {
public:
  explicit Writer(std::ostream& out);
  void Write(const std::vector<uint8_t>& bytes);

private:
  std::ostream& out_;
  // TODO: std::queue should be enough?
  std::deque<uint8_t> buf_;
};

class JSObject;

class Object {
public:
  enum class Type {
    Null,
    Undefined,
    Bool,
    Number,
    String,
    Object,
  };

  static Object Undefined();

  Object();
  explicit Object(bool b);
  explicit Object(double num);
  explicit Object(const std::string& str);
  explicit Object(const std::vector<uint8_t>& bytes);
  explicit Object(std::shared_ptr<JSObject> jsobject);
  explicit Object(const std::vector<Object>& array);

  Object(const Object& rhs);
  Object& operator=(const Object& rhs);
  bool operator<(const Object& rhs) const;

  bool IsNull() const;
  bool IsUndefined() const;
  bool IsBool() const;
  bool IsNumber() const;
  bool IsString() const;
  bool IsBytes() const;
  bool IsJSObject() const;
  bool IsArray() const;

  bool ToBool() const;
  double ToNumber() const;
  std::string ToString() const;
  std::vector<uint8_t>& ToBytes();
  const std::vector<uint8_t>& ToBytes() const;
  JSObject& ToJSObject();
  const JSObject& ToJSObject() const;
  std::vector<Object>& ToArray();

  std::string Inspect() const;

private:
  explicit Object(Type type);
  Object(Type type, double num);

  Type type_ = Type::Null;
  double num_value_ = 0;
  std::shared_ptr<std::vector<uint8_t>> bytes_value_;
  std::shared_ptr<JSObject> jsobject_value_;
  std::shared_ptr<std::vector<Object>> array_value_;
};

class JSObject {
public:
  class IValues {
  public:
    virtual ~IValues();
    virtual Object Get(const std::string& key) = 0;
    virtual void Set(const std::string& key, Object value) = 0;
    virtual void Remove(const std::string& key) = 0;
  };

  using JSFunc = std::function<Object (Object, std::vector<Object>)>;

  static Object Global();
  static std::shared_ptr<JSObject> Go(std::unique_ptr<IValues> values);
  static std::shared_ptr<JSObject> Enosys(const std::string& name);

  static Object ReflectGet(Object target, const std::string& key);
  static void ReflectSet(Object target, const std::string& key, Object value);
  static void ReflectDelete(Object target, const std::string& key);
  static Object ReflectConstruct(Object target, std::vector<Object> args);
  static Object ReflectApply(Object target, Object self, std::vector<Object> args);

  JSObject(const std::string& name);
  JSObject(const std::map<std::string, Object>& values);
  JSObject(std::unique_ptr<IValues> values);
  JSObject(const std::string& name, std::unique_ptr<IValues> values);
  JSObject(const std::string& name, const std::map<std::string, Object>& values);
  JSObject(JSFunc fn);
  JSObject(const std::string& name, std::unique_ptr<IValues> values, JSFunc fn, bool ctor);

  bool IsFunction() const;
  Object Get(const std::string& key);
  void Set(const std::string& key, Object value);
  void Delete(const std::string& key);
  Object Invoke(std::vector<Object> args);

  std::string ToString() const;

private:
  class DictionaryValues : public IValues {
  public:
    explicit DictionaryValues(const std::map<std::string, Object>& dict);
    Object Get(const std::string& key) override;
    void Set(const std::string& key, Object value) override;
    void Remove(const std::string& key) override;

  private:
    std::map<std::string, Object> dict_;
  };

  class FS {
  public:
    FS();
    Object Write(Object self, std::vector<Object> args);

  private:
    Writer stdout_;
    Writer stderr_;
  };

  static std::shared_ptr<JSObject> MakeGlobal();

  const std::string name_ = "(JSObject)";
  std::unique_ptr<IValues> values_ = nullptr;
  JSFunc fn_;
  const bool ctor_ = false;
};

}

#endif  // {{.IncludeGuard}}
`))

var jsCppTmpl = template.Must(template.New("js.cpp").Parse(`// Code generated by go2cpp. DO NOT EDIT.

#include "autogen/js.h"

#include <algorithm>
#include <cassert>
#include <cstdlib>
#include <random>
#include <tuple>

namespace {{.Namespace}} {

namespace {

void error(const std::string& msg) {
  std::cerr << msg << std::endl;
  assert(false);
  std::exit(1);
}

std::string JoinObjects(const std::vector<Object>& objs) {
  std::string str;
  for (int i = 0; i < objs.size(); i++) {
    str += objs[i].Inspect();
    if (i < objs.size() - 1) {
      str += ", ";
    }
  }
  return str;
}

void WriteObjects(std::ostream& out, const std::vector<Object>& objs) {
  std::vector<std::string> inspects(objs.size());
  for (int i = 0; i < objs.size(); i++) {
    out << objs[i].Inspect();
    if (i < objs.size() - 1) {
      out << ", ";
    }
  }
  out << std::endl;
}

}  // namespace

Writer::Writer(std::ostream& out)
    : out_{out} {
}

void Writer::Write(const std::vector<uint8_t>& bytes) {
  buf_.insert(buf_.end(), bytes.begin(), bytes.end());
  for (;;) {
    auto it = std::find(buf_.begin(), buf_.end(), '\n');
    if (it == buf_.end()) {
      break;
    }
    std::string str(buf_.begin(), it);
    out_ << str << std::endl;
    ++it;
    buf_.erase(buf_.begin(), it);
  }
}

Object Object::Undefined() {
  static Object& undefined = *new Object(Type::Undefined);
  return undefined;
}

Object::Object() = default;

Object::Object(bool b)
    : type_{Type::Bool},
      num_value_{static_cast<double>(b)}{
}

Object::Object(double num)
    : type_{Type::Number},
      num_value_{num} {
}

Object::Object(const std::string& str)
    : type_{Type::String},
      bytes_value_{std::make_shared<std::vector<uint8_t>>(str.begin(), str.end())} {
}

Object::Object(const std::vector<uint8_t>& bytes)
    : type_{Type::Object},
      bytes_value_{std::make_shared<std::vector<uint8_t>>(bytes.begin(), bytes.end())} {
}

Object::Object(std::shared_ptr<JSObject> jsobject)
    : type_{Type::Object},
      jsobject_value_{jsobject} {
}

Object::Object(const std::vector<Object>& array)
    : type_{Type::Object},
      array_value_{std::make_shared<std::vector<Object>>(array.begin(), array.end())} {
}

Object::Object(const Object& rhs) = default;

Object& Object::operator=(const Object& rhs) = default;

bool Object::operator<(const Object& rhs) const {
  return std::tie(type_, num_value_, bytes_value_, jsobject_value_, array_value_) <
      std::tie(rhs.type_, rhs.num_value_, rhs.bytes_value_, rhs.jsobject_value_, rhs.array_value_);
}

Object::Object(Type type)
    : type_{type} {
}

Object::Object(Type type, double num)
    : type_{type},
      num_value_{num} {
}

bool Object::IsNull() const {
  return type_ == Type::Null;
}

bool Object::IsUndefined() const {
  return type_ == Type::Undefined;
}

bool Object::IsBool() const {
  return type_ == Type::Bool;
}

bool Object::IsNumber() const {
  return type_ == Type::Number;
}

bool Object::IsString() const {
  return type_ == Type::String;
}

bool Object::IsBytes() const {
  return type_ == Type::Object && !jsobject_value_ && !array_value_;
}

bool Object::IsJSObject() const {
  return type_ == Type::Object && !!jsobject_value_;
}

bool Object::IsArray() const {
  return type_ == Type::Object && !!array_value_;
}

bool Object::ToBool() const {
  if (type_ != Type::Bool) {
    error("Object::ToBool: the type must be Type::Bool but not: " + Inspect());
  }
  return static_cast<bool>(num_value_);
}

double Object::ToNumber() const {
  if (type_ != Type::Number) {
    error("Object::ToNumber: the type must be Type::Number but not: " + Inspect());
  }
  return num_value_;
}

std::string Object::ToString() const {
  if (type_ != Type::String) {
    error("Object::ToString: the type must be Type::String but not: " + Inspect());
  }
  if (!bytes_value_) {
    error("Object::ToString: bytes_value_ must not be null");
  }
  return std::string(bytes_value_->begin(), bytes_value_->end());
}

std::vector<uint8_t>& Object::ToBytes() {
  if (type_ != Type::Object) {
    error("Object::ToBytes: the type must be Type::Object but not: " + Inspect());
  }
  if (!bytes_value_) {
    error("Object::ToBytes: bytes_value_ must not be null");
  }
  return *bytes_value_;
}

const std::vector<uint8_t>& Object::ToBytes() const {
  if (type_ != Type::Object) {
    error("Object::ToBytes: the type must be Type::Object but not: " + Inspect());
  }
  if (!bytes_value_) {
    error("Object::ToBytes: bytes_value_ must not be null");
  }
  return *bytes_value_;
}

JSObject& Object::ToJSObject() {
  if (type_ != Type::Object) {
    error("Object::ToJSObject: the type must be Type::Object but not: " + Inspect());
  }
  if (!jsobject_value_) {
    error("Object::ToJSObject: jsobject_value_ must not be null");
  }
  return *jsobject_value_;
}

const JSObject& Object::ToJSObject() const {
  if (type_ != Type::Object) {
    error("Object::ToJSObject: the type must be Type::Object but not: " + Inspect());
  }
  if (!jsobject_value_) {
    error("Object::ToJSObject: jsobject_value_ must not be null");
  }
  return *jsobject_value_;
}

std::vector<Object>& Object::ToArray() {
  if (type_ != Type::Object) {
    error("Object::ToArray: the type must be Type::Object but not: " + Inspect());
  }
  if (!array_value_) {
    error("Object::ToArray: array_value_ must not be null");
  }
  return *array_value_;
}

std::string Object::Inspect() const {
  switch (type_) {
  case Type::Null:
    return "null";
  case Type::Undefined:
    return "undefined";
  case Type::Bool:
    return ToBool() ? "true" : "false";
  case Type::Number:
    return std::to_string(ToNumber());
  case Type::String:
    return ToString();
  case Type::Object:
    if (IsJSObject()) {
      return ToJSObject().ToString();
    }
    return "(object)";
  default:
    error("invalid type: " + std::to_string(static_cast<int>(type_)));
  }
  return "";
}

JSObject::IValues::~IValues() = default;

JSObject::DictionaryValues::DictionaryValues(const std::map<std::string, Object>& dict)
    : dict_{dict} {
}

Object JSObject::DictionaryValues::Get(const std::string& key) {
  auto it = dict_.find(key);
  if (it == dict_.end()) {
    return Object{};
  }
  return it->second;
}

void JSObject::DictionaryValues::Set(const std::string& key, Object object) {
  dict_[key] = object;
}

void JSObject::DictionaryValues::Remove(const std::string& key) {
  dict_.erase(key);
}

JSObject::FS::FS()
    : stdout_{std::cout},
      stderr_{std::cerr} {
}

Object JSObject::FS::Write(Object self, std::vector<Object> args) {
  int fd = (int)(args[0].ToNumber());
  std::vector<uint8_t>& buf = args[1].ToBytes();
  int offset = (int)(args[2].ToNumber());
  int length = (int)(args[3].ToNumber());
  Object position = args[4];
  Object callback = args[5];
  if (offset != 0 || length != buf.size()) {
    ReflectApply(callback, Object{}, std::vector<Object>{ Object{Enosys("write")} });
    return Object{};
  }
  if (!position.IsNull()) {
    ReflectApply(callback, Object{}, std::vector<Object>{ Object{Enosys("write")} });
    return Object{};
  }
  switch (fd) {
  case 1:
    stdout_.Write(buf);
    break;
  case 2:
    stderr_.Write(buf);
    break;
  default:
    ReflectApply(callback, Object{}, std::vector<Object>{ Object{Enosys("write")} });
    break;
  }
  ReflectApply(callback, Object{}, std::vector<Object>{ Object{}, Object{static_cast<double>(buf.size())} });
  return Object{};
}

Object JSObject::Global() {
  static Object& global = *new Object(MakeGlobal());
  return global;
}

std::shared_ptr<JSObject> JSObject::MakeGlobal() {
  std::shared_ptr<JSObject> arr = std::make_shared<JSObject>("Array");
  std::shared_ptr<JSObject> obj = std::make_shared<JSObject>("Object");
  std::shared_ptr<JSObject> u8 = std::make_shared<JSObject>("Uint8Array", nullptr,
    [](Object self, std::vector<Object> args) -> Object {
      if (args.size() == 0) {
        return Object{std::vector<uint8_t>{}};
      }
      if (args.size() == 1) {
        Object len = args[0];
        if (len.IsNumber()) {
          return Object{std::vector<uint8_t>(static_cast<int>(len.ToNumber()))};
        }
        error("new Uint8Array(" + args[0].Inspect() + ") is not implemented");
      }
      error("new Uint8Array with " + std::to_string(args.size()) + " args is not implemented");
      return Object{};
    }, true);

  Object getRandomValues{std::make_shared<JSObject>(
    [](Object self, std::vector<Object> args) -> Object {
      std::vector<uint8_t>& bs = args[0].ToBytes();
      // TODO: Use cryptographically strong random values instead of std::random_device.
      static std::random_device rd;
      std::uniform_int_distribution<uint8_t> dist(0, 255);
      for (int i = 0; i < bs.size(); i++) {
        bs[i] = dist(rd);
      }
      return Object{};
    })};
  std::shared_ptr<JSObject> crypto = std::make_shared<JSObject>("crypto", std::map<std::string, Object>{
    {"getRandomValues", getRandomValues},
  });

  static Object& writeObjectsToStdout = *new Object(std::make_shared<JSObject>(
    [](Object self, std::vector<Object> args) -> Object {
      WriteObjects(std::cout, args);
      return Object{};
    }));
  static Object& writeObjectsToStderr = *new Object(std::make_shared<JSObject>(
    [](Object self, std::vector<Object> args) -> Object {
      WriteObjects(std::cerr, args);
      return Object{};
    }));
  std::shared_ptr<JSObject> console = std::make_shared<JSObject>("console", std::map<std::string, Object>{
    {"error", writeObjectsToStderr},
    {"debug", writeObjectsToStderr},
    {"info", writeObjectsToStdout},
    {"log", writeObjectsToStdout},
    {"warm", writeObjectsToStderr},
  });

  std::shared_ptr<JSObject> fetch = std::make_shared<JSObject>(
    [](Object self, std::vector<Object> args) -> Object {
      // TODO: Implement this.
      return Object{};
    });

  static FS& fsimpl = *new FS();
  std::shared_ptr<JSObject> fs = std::make_shared<JSObject>("fs", std::map<std::string, Object>{
    {"constants", Object{std::make_shared<JSObject>(std::map<std::string, Object>{
        {"O_WRONLY", Object{-1.0}},
        {"O_RDWR", Object{-1.0}},
        {"O_CREAT", Object{-1.0}},
        {"O_TRUNC", Object{-1.0}},
        {"O_APPEND", Object{-1.0}},
        {"O_EXCL", Object{-1.0}},
      })}},
    {"write", Object{std::make_shared<JSObject>(
      [](Object self, std::vector<Object> args) -> Object {
        return fsimpl.Write(self, args);
      })}},
  });

  std::shared_ptr<JSObject> process = std::make_shared<JSObject>("process", std::map<std::string, Object>{
    {"pid", Object{-1.0}},
    {"ppid", Object{-1.0}},
  });

  std::shared_ptr<JSObject> global = std::make_shared<JSObject>("global", std::map<std::string, Object>{
    {"Array", Object{arr}},
    {"Object", Object{obj}},
    {"Uint8Array", Object{u8}},
    {"console", Object{console}},
    {"crypto", Object{crypto}},
    {"fetch", Object{fetch}},
    {"fs", Object{fs}},
    {"process", Object{process}},
  });
  return global;
}

std::shared_ptr<JSObject> JSObject::Go(std::unique_ptr<IValues> values) {
  return std::make_shared<JSObject>("go", std::move(values));
}

std::shared_ptr<JSObject> JSObject::Enosys(const std::string& name) {
  return std::make_shared<JSObject>(std::map<std::string, Object>{
    {"message", Object{name + " not implemented"}},
    {"code", Object{"ENOSYS"}},
  });
}

Object JSObject::ReflectGet(Object target, const std::string& key) {
  if (target.IsUndefined()) {
    error("get on undefined (key: " + key + ") is forbidden");
    return Object{};
  }
  if (target.IsNull()) {
    error("get on null (key: " + key + ") is forbidden");
    return Object{};
  }
  if (target.IsJSObject()) {
    return target.ToJSObject().Get(key);
  }
  if (target.IsArray()) {
    int idx = std::stoi(key);
    if (idx > 0 || (idx == 0 && key == "0")) {
      return target.ToArray()[idx];
    }
  }
  error(target.Inspect() + "." + key + " not found");
  return Object{};
}

void JSObject::ReflectSet(Object target, const std::string& key, Object value) {
  if (target.IsUndefined()) {
    error("set on undefined (key: " + key + ") is forbidden");
  }
  if (target.IsNull()) {
    error("set on null (key: " + key + ") is forbidden");
  }
  if (target.IsJSObject()) {
    target.ToJSObject().Set(key, value);
    return;
  }
  error(target.Inspect() + "." + key + " cannot be set");
}

void JSObject::ReflectDelete(Object target, const std::string& key) {
  if (target.IsUndefined()) {
    error("delete on undefined (key: " + key + ") is forbidden");
  }
  if (target.IsNull()) {
    error("delete on null (key: " + key + ") is forbidden");
  }
  if (target.IsJSObject()) {
    target.ToJSObject().Delete(key);
    return;
  }
  error(target.Inspect() + "." + key + " cannot be deleted");
}

Object JSObject::ReflectConstruct(Object target, std::vector<Object> args) {
  if (target.IsUndefined()) {
    error("new on undefined is forbidden");
    return Object{};
  }
  if (target.IsNull()) {
    error("new on null is forbidden");
    return Object{};
  }
  if (target.IsJSObject()) {
    JSObject& t = target.ToJSObject();
    if (!t.ctor_) {
      error(t.ToString() + " is not a constructor");
      return Object{};
    }
    return t.fn_(target, args);
  }
  error("new " + target.Inspect() + "(" + JoinObjects(args) + ") cannot be called");
  return Object{};
}

Object JSObject::ReflectApply(Object target, Object self, std::vector<Object> args) {
  if (target.IsUndefined()) {
    error("apply on undefined is forbidden");
    return Object{};
  }
  if (target.IsNull()) {
    error("apply on null is forbidden");
    return Object{};
  }
  if (target.IsJSObject()) {
    JSObject& t = target.ToJSObject();
    if (t.ctor_) {
      error(t.ToString() + " is a constructor");
      return Object{};
    }
    return t.fn_(self, args);
  }
  error(target.Inspect() + "(" + JoinObjects(args) + ") cannot be called");
  return Object{};
}

JSObject::JSObject(const std::string& name)
    : name_{name} {
}

JSObject::JSObject(const std::map<std::string, Object>& values)
    : values_{std::make_unique<DictionaryValues>(values)} {
}

JSObject::JSObject(std::unique_ptr<IValues> values)
    : values_{std::move(values)} {
}

JSObject::JSObject(const std::string& name, std::unique_ptr<IValues> values)
    : name_{name},
      values_{std::move(values)} {
}

JSObject::JSObject(const std::string& name, const std::map<std::string, Object>& values)
    : name_{name},
      values_{std::make_unique<DictionaryValues>(values)} {
}

JSObject::JSObject(JSFunc fn)
    : fn_{fn} {
}

JSObject::JSObject(const std::string& name, std::unique_ptr<IValues> values, JSFunc fn, bool ctor)
    : name_{name},
      values_{std::move(values)},
      fn_{fn},
      ctor_{ctor} {
}

bool JSObject::IsFunction() const {
  return !!fn_;
}

Object JSObject::Get(const std::string& key) {
  if (!values_) {
    error(ToString() + "." + key + " not found");
  }
  return values_->Get(key);
}

void JSObject::Set(const std::string& key, Object value) {
  if (!values_) {
    values_ = std::make_unique<DictionaryValues>(std::map<std::string, Object>{});
  }
  values_->Set(key, value);
}

void JSObject::Delete(const std::string& key) {
  if (!values_) {
    return;
  }
  values_->Remove(key);
}

Object JSObject::Invoke(std::vector<Object> args) {
  if (!fn_) {
    error(ToString() + " is not invokable since " + ToString() + " is not a function");
    return Object{};
  }
  if (ctor_) {
    error(ToString() + " is not invokable since " + ToString() + " is a constructor");
    return Object{};
  }
  return fn_(Object{}, args);
}

std::string JSObject::ToString() const {
  return name_;
}

}
`))
