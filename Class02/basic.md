# Basic

## k8s 依賴的 glog

k8s -> glog -> init -> flag parse parameter

## 返回值

- 多返回值
  - 函數返回任意數量的返回值
  - 多返回值的應用場景？錯誤處理
- 命名返回值
  - Go 的返回值可被命名，他們會被視作定義在函數頂部的變量
  - 返回值的名稱應當具有一定的意義，它可以作為文檔使用
  - 沒有參數的 return 語句返回已命名的參數

## 傳遞變長的參數

```go
func append(slice []Type, elems ...Type) []Type

func x() {
    myArray := []string{}
    myArray := append(myArray, "a","b","c")
}
```

## 回調函數 (Callback)

定義一個 Callback 函數，把它作為變量傳到一個函數裡去。

```go
func DoOperation(y int, f func(int,int)) {
    f(y,1)
}
```

## 閉包

- 匿名函數
  - 不能獨立存在
  - 可以複製給變量
    - x := func(){}
  - 可以直接調用
    - func (x, y int){println(x+y)}(1,2)
  - 可以作為函數返回值

使用場景: 想表達邏輯，又不想給函數命名

## defer

通常用來處理資源回收的

正常情況：

```go
func x() {
    l.Lock()
    defer l.Unlock()
    
    xxx,err := bbb()
    if err != nil {
        return
    }
}
```

鎖資源沒被回收情況：

```go
func x() {
    l.Lock()
    xxx,err := bbb()
    if err != nil {
        // Will not release the lock.
        return
    }
    l.Unlock()
}
```

## CSP: Communicating Sequential Process

- CSP: 描述兩個獨立的併發實體，通過共享的通信 channel 進行通信的併發模型
- Goroutine：輕量級線程
- 通道 Channel：攜程之間的通訊和同步。攜程之間解耦，但是和 channel 耦合

## Context

- 超市、取消操作或者一些異常情況，往往需要進行強佔式操作或者中斷後續操作。
- Context 是設置截止日期、同步信號，傳遞請求相關值的結構體。

```go
type Context interface {
    Deadline() (deadline time.Time, ok bool)
    Done() <-chan struct{}
    Err() error
    Value(key interface{}) interface{}
}
```


