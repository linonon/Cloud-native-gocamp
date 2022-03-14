# GC

## Heap 內存管理

- (Heap<-)Allocator：內存分配器，處理動態內存分配請求
- Mutator(->Allocator)：用戶程序，通過 Allocator 創建對象
- (Colletor<-)Object Header(->Allocator)：Collector 和 Allocator 同步對象元數據
- Collector(->Heap)： 垃圾回收器，回收內存空間。

Heap 內存管理
![Heap 內存管理](pic/Heap-manage.png)

## ThreadCacheMalloc

TCMalloc概覽
![TCMalloc概覽](pic/TCMalloc.png)

- page: 8K
- span: 內存塊，一個或多個 page
- sizeclass：span 的 size class
- object：對象，假設 obj = 16b，span = 8K，那 span 就能分成 8k/16b = 512個 obj，分配的話就分配 1 個 obj 出去

- obj 大小定義
  - 小： 0 ～ 256k
  - 中： 256k ～ 1M
  - 大： > 1M
- 小的分配流程：
  - TC -> CentralCache -> HeapPage，大部分時候，TC緩存都是足夠的，不需要走向下層。無系統調用配合無鎖分配，所以分配效率非常高。
- 中：直接在 PageHeap中選適當的大小。128 Page = 1M
- 大： 從 large span set 選擇合適數量的頁面組成 span，用來存儲數據。

## Go語言的內存分配

Go語言的內存分配
![Go語言的內存分配](pic/Go-mem.png)

重點：兩個大小一樣的Span Class 對應一個 size class，一個存指針，一個存直接引用，直接引用的span無需內存回收。