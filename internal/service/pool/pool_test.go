package pool_test

import (
	"testing"

	"github.com/Ko4etov/go-metrics/internal/service/pool"
)

type TestStruct struct {
	value int
	data  []byte
	count int
}

func (ts *TestStruct) Reset() {
	ts.value = 0
	ts.data = ts.data[:0]
	ts.count = 0
}

func TestPoolNew(t *testing.T) {
	p := pool.New(func() *TestStruct {
		return &TestStruct{
			value: 42,
			data:  make([]byte, 0, 100),
			count: 100,
		}
	})

	if p == nil {
		t.Fatal("New should return non-nil pool")
	}
}

func TestPoolGetPut(t *testing.T) {
	creationCount := 0
	p := pool.New(func() *TestStruct {
		creationCount++
		return &TestStruct{
			value: creationCount * 10,
			data:  make([]byte, 0, 10),
			count: creationCount,
		}
	})

	obj1 := p.Get()
	if obj1 == nil {
		t.Fatal("Get should return non-nil object")
	}
	if creationCount != 1 {
		t.Errorf("Expected creationCount = 1, got %d", creationCount)
	}
	if obj1.value != 10 || obj1.count != 1 {
		t.Errorf("New object should have initialized values, got value=%d, count=%d", obj1.value, obj1.count)
	}

	// Изменяем объект
	obj1.value = 999
	obj1.data = append(obj1.data, 1, 2, 3)
	obj1.count = 777

	p.Put(obj1)

	obj2 := p.Get()
	if obj2 == nil {
		t.Fatal("Get should return non-nil object")
	}

	if obj2.value != 0 {
		t.Errorf("Reset object should have value=0, got %d", obj2.value)
	}
	if len(obj2.data) != 0 {
		t.Errorf("Reset object should have empty slice, got len=%d", len(obj2.data))
	}
	if obj2.count != 0 {
		t.Errorf("Reset object should have count=0, got %d", obj2.count)
	}
}

func TestPoolReuse(t *testing.T) {
	created := []*TestStruct{}
	p := pool.New(func() *TestStruct {
		obj := &TestStruct{value: len(created)}
		created = append(created, obj)
		return obj
	})

	objs := make([]*TestStruct, 5)
	for i := 0; i < 5; i++ {
		objs[i] = p.Get()
		objs[i].value = i + 100
	}

	if len(created) != 5 {
		t.Errorf("Expected 5 objects created, got %d", len(created))
	}

	for _, obj := range objs {
		p.Put(obj)
	}

	reused := make([]*TestStruct, 5)
	for i := 0; i < 5; i++ {
		reused[i] = p.Get()
	}

	if len(created) != 5 {
		t.Errorf("Expected no new objects created after reuse, still %d", len(created))
	}

	for i, obj := range reused {
		if obj.value != 0 {
			t.Errorf("Object %d should be reset to value=0, got %d", i, obj.value)
		}
	}
}

func TestPoolConcurrent(t *testing.T) {
	const goroutines = 10
	const iterations = 100

	p := pool.New(func() *TestStruct {
		return &TestStruct{}
	})

	done := make(chan bool)
	for g := 0; g < goroutines; g++ {
		go func(id int) {
			for i := 0; i < iterations; i++ {
				obj := p.Get()

				obj.value = id*1000 + i
				obj.data = append(obj.data, byte(i))
				obj.count++

				p.Put(obj)
			}
			done <- true
		}(g)
	}

	for g := 0; g < goroutines; g++ {
		<-done
	}

	obj := p.Get()
	defer p.Put(obj)

	if obj.value != 0 || len(obj.data) != 0 || obj.count != 0 {
		t.Errorf("Object should be properly reset after concurrent use")
	}
}