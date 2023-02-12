package snowflake

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

var ErrCurrentTime = errors.New("snowflake: current time error")

// 起始时间戳，用于用当前时间戳减去这个时间戳，算出偏移量
const twepoch int64 = 1288834974657

// workID占用的比特数
const workerIdBits int64 = 10

// 最大能够分配的workerId =1023
const maxWorkerId int64 = -1 ^ (-1 << workerIdBits)

// 自增序列号
const sequenceBits int64 = 12

// workID左移位数为自增序列号的位数
const workerIdShift int64 = sequenceBits

// 时间戳的左移位数为 自增序列号的位数+workID的位数
const timestampLeftShift int64 = sequenceBits + workerIdBits

// 后12位都为1
const sequenceMask int64 = -1 ^ (-1 << sequenceBits)

type Snowflake struct {
	// 保存该节点的workId
	workerId int64
	// 序列号
	sequence int64
	// 上一次请求id时所用的时间戳
	lastTimestamp int64
	// 锁
	lock sync.Mutex
}

// NewSnowflake 创建实例
func NewSnowflake(workerId int64) *Snowflake {
	return &Snowflake{
		workerId:      workerId,
		sequence:      0,
		lastTimestamp: -1,
	}
}

func MaxWorkerId() int64 {
	return maxWorkerId
}

// GetId 获取id
//
// 生成id号需要的时间戳和序列号
// 1. 时间戳要求大于等于上一次用的时间戳（主要解决机器工作时NTP时间同步问题）
// 2. 序列号在时间戳相等的情况下要递增，大于的情况下回到起点
func (s *Snowflake) GetId() (int64, error) {
	rand.Seed(time.Now().UnixNano())
	s.lock.Lock()
	defer s.lock.Unlock()
	// 获取当前时间戳，timestamp用于记录生成id的时间戳
	timestamp := timeGen()
	// 如果比上一次记录的时间戳早，也就是NTP造成时间回退了
	if timestamp < s.lastTimestamp {
		offset := s.lastTimestamp - timestamp
		if offset <= 5 {
			// 等待 2*offset ms就可以唤醒重新尝试获取锁继续执行。当然，在这段时间内lastTimestamp很可能又被更新了
			time.Sleep(time.Duration(offset<<1) * time.Millisecond)
			// 重新获取当前时间戳，理论上这次应该比上一次记录的时间戳迟了
			timestamp = timeGen()
			// 如果还是早，这绝对是有问题的
			if timestamp < s.lastTimestamp {
				return 0, ErrCurrentTime
			}
		} else {
			return 0, ErrCurrentTime
		}
	}
	// 如果从上一个逻辑分支产生的timestamp仍然和lastTimestamp相等
	if s.lastTimestamp == timestamp {
		// 自增序列+1然后取后12位的值
		s.sequence = (s.sequence + 1) & sequenceMask
		// seq 为0的时候表示当前毫秒12位自增序列用完了，应该用下一毫秒时间来区别，否则就重复了
		if s.sequence == 0 {
			// 对seq做随机作为起始，主要出于DB分表均匀的考虑
			s.sequence = int64(rand.Int31n(100))
			// 生成比lastTimestamp滞后的时间戳
			timestamp = tilNextMillis(s.lastTimestamp)
		}
	} else {
		// 如果是新的ms开始，序列号要重新回到大致的起点
		s.sequence = rand.Int63n(100)
	}
	// 记录这次请求id的时间戳，用于下一个请求进行比较
	s.lastTimestamp = timestamp
	// 利用生成的时间戳、序列号和workId组合成id
	id := ((timestamp - twepoch) << timestampLeftShift) | (s.workerId << workerIdShift) | s.sequence
	return id, nil
}

func tilNextMillis(lastTimestamp int64) int64 {
	timestamp := timeGen()
	for timestamp <= lastTimestamp {
		time.Sleep(100 * time.Microsecond)
		timestamp = timeGen()
	}
	return timestamp
}

func timeGen() int64 {
	return time.Now().UnixMilli()
}
