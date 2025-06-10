package suit

import (
	"container/heap"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type result struct {
	index   int
	md      string
	summary string
}

/*
parallelProcessing 并行处理文本分块并按顺序返回结果

@param chunks 需要处理的文本分块切片
@return 处理完成并按优先级排序的结果切片

工作流程：
1. 创建固定大小的worker池处理任务
2. 将任务索引分发给空闲worker
3. 收集所有worker返回的结果
4. 使用优先队列对结果进行排序
*/
func (tp *TextProcessor) parallelProcessing(chunks []string) []result {
	totalTasks := len(chunks)
	var progress atomic.Int64
	progress.Store(0)
	progressChan := make(chan int64, 100)
	defer close(progressChan)

	// 启动进度监控协程
	progressDone := make(chan struct{})
	defer close(progressDone)
	go monitorProgress(&progress, totalTasks, progressDone)

	var wg sync.WaitGroup
	results := make(chan result, len(chunks))
	workQueue := make(chan int, tp.maxWorkers)

	/*
		启动固定数量的worker协程池
		每个worker持续从workQueue获取任务索引
		处理完成后通过results通道返回结果
	*/
	for range tp.maxWorkers {
		go tp.worker(chunks, results, &wg, workQueue, &progress)
	}

	/*
		任务分发阶段：
		1. 为每个任务块增加WaitGroup计数
		2. 将任务索引发送到工作队列
		3. 关闭工作队列表示任务分发完成
	*/
	for i := range chunks {
		wg.Add(1)
		workQueue <- i
	}
	close(workQueue)
	wg.Wait()
	close(results)

	/*
		结果收集与排序：
		1. 将所有结果存入优先队列
		2. 按优先级顺序弹出结果
		3. 转换为有序切片返回
	*/
	var sortedResults priorityQueue
	for res := range results {
		heap.Push(&sortedResults, res)
	}

	final := make([]result, len(chunks))
	for i := 0; sortedResults.Len() > 0; i++ {
		final[i] = heap.Pop(&sortedResults).(result)
	}
	return final
}

// worker 是TextProcessor的工作协程，负责并发处理文本块并生成结果。
// 参数：
//
//	chunks: 分片后的文本块切片，每个元素为待处理的文本字符串
//	results: 单向发送通道，用于将处理结果（包含索引、markdown内容和摘要）发送回主协程
//	wg: 同步等待组，用于协调多个worker的完成状态
//	queue: 只读通道，提供待处理文本块的索引序列
//
// 该方法通过循环处理分配给它的文本块索引，处理完成后：
// 1. 通过results通道发送结果
// 2. 调用wg.Done()通知任务完成
func (tp *TextProcessor) worker(chunks []string, results chan<- result, wg *sync.WaitGroup, queue <-chan int, progress *atomic.Int64) {
	// 从任务队列接收索引并处理：
	// - 获取上下文信息
	// - 处理文本块生成markdown和摘要
	// - 发送结果到results通道
	// - 通知等待组任务完成
	for i := range queue {
		context := getContext(chunks, i)
		md, summary := tp.processChunk(i, chunks[i], context)
		results <- result{index: i, md: md, summary: summary}
		progress.Add(1)
		wg.Done()
	}
}

func getContext(chunks []string, currentIdx int) string {
	start := max(currentIdx-3, 0)
	return strings.Join(chunks[start:currentIdx], "\n")
}

// 实现优先队列
type priorityQueue []result

func (pq priorityQueue) Len() int           { return len(pq) }
func (pq priorityQueue) Less(i, j int) bool { return pq[i].index < pq[j].index }
func (pq priorityQueue) Swap(i, j int)      { pq[i], pq[j] = pq[j], pq[i] }
func (pq *priorityQueue) Push(x any)        { *pq = append(*pq, x.(result)) }
func (pq *priorityQueue) Pop() any {
	old := *pq
	n := len(old)
	x := old[n-1]
	*pq = old[0 : n-1]
	return x
}

func monitorProgress(progress *atomic.Int64, totalTasks int, done <-chan struct{}) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			current := progress.Load()
			percent := float64(current) / float64(totalTasks) * 100
			fmt.Printf("\r处理进度: %d/%d (%.1f%%) ", current, totalTasks, percent)
		case <-done:
			fmt.Println("\n进度: 100% 完成")
			return
		}
	}
}
