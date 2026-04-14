
# 🚀 dial — Lightweight Go HTTP Router & Client

A minimal and powerful Go library for building internal microservices with support for:

* 🔌 **Unix Socket communication**
* 🌐 **TCP transport**
* 🧭 Clean routing via `gorilla/mux`
* ⚡ Simple, generic client API

---

## 📦 Installation

```bash
go get github.com/saravanan611/dial
```

---

## ⚡ Quick Start

### 🖥️ Server

```go
package main

import (
    "net/http"

    "github.com/saravanan611/dial/dial"
    "github.com/saravanan611/log"
)

func main() {
    log.EnableInfo()

    router := dial.NewRouter()

    // 🔗 Routes
    router.HandleFunc("/sum", Sum).SetHrdKey("sample").Methods(http.MethodPost)
    router.HandleFunc("/sub", Sub).Methods(http.MethodPost)

    // 🚀 Start server (Unix socket → temp/sample.sock)
    if err := router.Start("unix", "sample"); err != nil {
        log.Err(err)
    }
}

type NumStruct struct {
    Num1 int `json:"num1"`
    Num2 int `json:"num2"`
}

func Sum(pResp *dial.Resp, pReq *dial.Request) {
    var nums NumStruct
    if err := pReq.Read(&nums); err != nil {
        pResp.SendError("SUM01", err)
        return
    }
    dial.Send[int](pResp, nums.Num1+nums.Num2)
}

func Sub(pResp *dial.Resp, pReq *dial.Request) {
    var nums NumStruct
    if err := pReq.Read(&nums); err != nil {
        pResp.SendError("SUB01", err)
        return
    }
    dial.Send[int](pResp, nums.Num1-nums.Num2)
}
```

---

### 📡 Client

```go
package main

import (
    "net/http"

    "github.com/saravanan611/dial/dial"
    "github.com/saravanan611/log"
)

type ReqStruct struct {
    Num1 int `json:"num1"`
    Num2 int `json:"num2"`
}

func main() {
    log.EnableInfo()

    req := ReqStruct{Num1: 10, Num2: 20}
    headers := map[string]string{"sample": "hi this is a sample header"}

    // ➕ Sum
    sum, _ := dial.Call[ReqStruct, int](
        "http://sample.localhost/sum",
        http.MethodPost,
        headers,
        req,
    )

    // ➖ Sub
    sub, _ := dial.Call[ReqStruct, int](
        "http://sample.localhost/sub",
        http.MethodPost,
        headers,
        req,
    )

    log.Info("Sum: %d", sum) // 30
    log.Info("Sub: %d", sub) // -10
}
```

---

## 🧠 Smart Unix Socket Routing

> 🧩 `http://<name>.localhost/...` → Automatically maps to:

```
temp/<name>.sock
```

✔ No config needed
✔ Clean service-to-service calls

---

## 📚 API Reference

### 🧭 Router

#### 🆕 `dial.NewRouter()`

Creates a router with built-in:

* 🆔 Request ID generation
* 🌍 CORS support
* 🔁 OPTIONS handling

---

#### 🔗 `HandleFunc(path, handler)`

Register routes easily:

```go
router.HandleFunc("/path", handler)
```

---

#### 🚦 `Methods(...)`

Restrict HTTP methods:

```go
.Methods(http.MethodGet, http.MethodPost)
```

✔ `OPTIONS` is auto-added

---

#### 🧾 `SetHrdKey(...)`

Add custom headers for CORS:

```go
.SetHrdKey("Authorization", "X-Custom")
```

---

#### ▶️ `Start(type, name)`

Start server:

| Type 🧩  | Example    | Behavior           |
| -------- | ---------- | ------------------ |
| `"unix"` | `"sample"` | `temp/sample.sock` |
| `"tcp"`  | `":8080"`  | Port binding       |

---

## 🛠️ Handler Utilities

### 📥 Read Request

```go
pReq.Read(&dest)
```

✔ Supports:

* JSON 🧾
* Proto ⚡

---

### 📤 Send Response

```go
dial.Send[T](pResp, data)
```

---

### ❌ Send Error

```go
pResp.SendError("ERR01", err)
```

✔ Structured error response

---

## 📡 Client API

### 🔄 `dial.Call`

```go
result, err := dial.Call[Req, Resp](
    url,
    method,
    headers,
    body,
)
```

#### ✨ Features

* 🔁 Auto serialization/deserialization
* 🔌 Unix socket routing (`*.localhost`)
* 🧩 Generic types

---

## 🌍 Global Configuration

Call once at startup:

| Function ⚙️             | Purpose                   |
| ----------------------- | ------------------------- |
| `SetOrgin(...)`         | Allowed CORS origins      |
| `EnableCred()`          | Enable credentials        |
| `SetOrginCheckFunc(fn)` | Dynamic origin validation |

---

## 🚚 Transport Modes

| URL 🌐                         | Transport 🔌 | Path 📁             |
| ------------------------------ | ------------ | ------------------- |
| `http://service.localhost/...` | Unix Socket  | `temp/service.sock` |
| `http://host:port/...`         | TCP          | Direct              |

---

## 💡 Why dial?

* ⚡ Faster internal communication (Unix sockets)
* 🧼 Clean and minimal API
* 🔧 Zero-config service discovery
* 🧩 Strong typing with generics
* 🚀 Perfect for microservices

---
