package pool

import (
	"sync"
	"testing"
)

// testStruct - тестовая структура, реализующая Resetter через pointer receiver
type testStruct struct {
	value      int
	text       string
	slice      []int
	data       map[string]int
	resetCount int
}

// Reset implements the Resetter interface with a pointer receiver
func (ts *testStruct) Reset() {
	ts.value = 0
	ts.text = ""
	if ts.slice != nil {
		ts.slice = ts.slice[:0]
	}
	if ts.data != nil {
		clear(ts.data)
	}
	ts.resetCount++
}

type anotherStruct struct {
	flag bool
	data []string
}

func (a *anotherStruct) Reset() {
	a.flag = false
	a.data = a.data[:0]
}

// TestNew проверяет создание нового пула
func TestNew(t *testing.T) {
	pool := New[*testStruct]()

	if pool == nil {
		t.Fatal("New() returned nil")
	}

	if pool.internal == nil {
		t.Error("Internal sync.Pool should not be nil")
	}
}

// TestGetPut проверяет базовые операции Get и Put
func TestGetPut(t *testing.T) {
	pool := New[*testStruct]()

	// Получаем объект из пула
	obj := pool.Get()

	// Если объект nil, создаем его (это может случиться для интерфейсных типов)
	if obj == nil {
		obj = &testStruct{}
	}

	// Изменяем состояние объекта
	obj.value = 42
	obj.text = "hello"
	obj.slice = append(obj.slice, 1, 2, 3)
	obj.data = map[string]int{"key": 123}

	initialResetCount := obj.resetCount

	// Возвращаем объект в пул
	pool.Put(obj)

	// Получаем объект снова
	obj = pool.Get()
	if obj == nil {
		obj = &testStruct{}
	}

	// Проверяем, что Reset был вызван
	if obj.resetCount != initialResetCount+1 {
		t.Errorf("Reset() was not called. Expected resetCount %d, got %d",
			initialResetCount+1, obj.resetCount)
	}

	// Проверяем, что состояние сброшено
	if obj.value != 0 {
		t.Errorf("Value was not reset. Expected 0, got %d", obj.value)
	}
	if obj.text != "" {
		t.Errorf("Text was not reset. Expected empty string, got %q", obj.text)
	}
	if len(obj.slice) != 0 {
		t.Errorf("Slice was not reset. Expected length 0, got %d", len(obj.slice))
	}
	if len(obj.data) != 0 {
		t.Errorf("Map was not reset. Expected length 0, got %d", len(obj.data))
	}
}

// TestGetNewObject проверяет, что Get возвращает новый объект когда пул пуст
func TestGetNewObject(t *testing.T) {
	pool := New[*testStruct]()

	obj1 := pool.Get()
	if obj1 == nil {
		obj1 = &testStruct{}
	}

	obj2 := pool.Get()
	if obj2 == nil {
		obj2 = &testStruct{}
	}

	// Это должны быть разные объекты
	if obj1 == obj2 {
		t.Error("Get() should return different objects when pool is empty")
	}
}

// TestReuse проверяет повторное использование объектов
func TestReuse(t *testing.T) {
	pool := New[*testStruct]()

	// Получаем объект и возвращаем его
	obj := pool.Get()
	if obj == nil {
		obj = &testStruct{}
	}

	obj.value = 100
	originalAddr := obj
	pool.Put(obj)

	// Получаем объект снова - это должен быть тот же объект
	reusedObj := pool.Get()
	if reusedObj == nil {
		reusedObj = &testStruct{}
	}

	// Проверяем, что это тот же объект
	if originalAddr != reusedObj {
		t.Error("Pool should reuse objects after Put")
	}

	// Проверяем, что состояние сброшено
	if reusedObj.value != 0 {
		t.Errorf("Reused object was not reset. Expected value 0, got %d", reusedObj.value)
	}
}

// TestConcurrent использование пула в конкурентной среде
func TestConcurrent(t *testing.T) {
	pool := New[*testStruct]()

	var wg sync.WaitGroup
	iterations := 100

	// Запускаем несколько горутин
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Получаем объект
			obj := pool.Get()
			if obj == nil {
				obj = &testStruct{}
			}

			// Используем объект
			obj.value = id
			obj.text = "test"

			// Возвращаем в пул
			pool.Put(obj)
		}(i)
	}

	wg.Wait()

	// Проверяем, что пул работает корректно
	obj := pool.Get()
	if obj == nil {
		obj = &testStruct{}
	}

	if obj.value != 0 {
		t.Error("Pool should work correctly under concurrent load")
	}
}

// TestResetCalled проверяет, что Reset вызывается при каждом Put
func TestResetCalled(t *testing.T) {
	pool := New[*testStruct]()

	obj := pool.Get()
	if obj == nil {
		obj = &testStruct{}
	}

	initialResetCount := obj.resetCount

	// Многократно используем и возвращаем объект
	for i := 0; i < 5; i++ {
		obj.value = i * 10
		pool.Put(obj)
		obj = pool.Get()
		if obj == nil {
			obj = &testStruct{}
		}
	}

	// Reset должен вызываться при каждом Put
	expectedResetCount := initialResetCount + 5
	if obj.resetCount != expectedResetCount {
		t.Errorf("Reset should be called on every Put. Expected %d, got %d",
			expectedResetCount, obj.resetCount)
	}
}

// TestWithPointerType проверяет работу с указателями на разные структуры
func TestWithPointerType(t *testing.T) {
	pool := New[*anotherStruct]()
	obj := pool.Get()
	if obj == nil {
		obj = &anotherStruct{}
	}
	obj.flag = true
	obj.data = append(obj.data, "test")
	pool.Put(obj)

	objAgain := pool.Get()
	if objAgain == nil {
		objAgain = &anotherStruct{}
	}
	if objAgain.flag != false {
		t.Error("Reset should work for pointer types implementing Resetter")
	}
	if len(objAgain.data) != 0 {
		t.Error("Slice should be reset in pointer types")
	}
}

// BenchmarkPool измеряет производительность пула
func BenchmarkPool(b *testing.B) {
	pool := New[*testStruct]()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := pool.Get()
			if obj == nil {
				obj = &testStruct{}
			}
			// Симуляция работы с объектом
			obj.value = 42
			obj.text = "benchmark"
			pool.Put(obj)
		}
	})
}
