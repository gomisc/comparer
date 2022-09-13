package comparer

import (
	"reflect"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// IReporter - интерфейс репротера
type IReporter interface {
	// CustomReport - возвращает отчёт
	CustomReport() string
	// PushStep - метод необходимый для `cmp.Reporter`
	PushStep(cmp.PathStep)
	// Report - метод необходимый для `cmp.Reporter`
	Report(cmp.Result)
	// PopStep - метод необходимый для `cmp.Reporter`
	PopStep()
}

// Comparer - интерфейс сравнения объектов
type Comparer interface {
	WithIgnoreUnexportedOf(i ...interface{}) Comparer
	WithIgnoreEmptySlices() Comparer
	WithSortSlices(lessFunc LessSliсes) Comparer
	WithAllowUnexportedOf(i ...interface{}) Comparer
	WithIgnoreFields(typ interface{}, names ...string) Comparer
	WithCustomReporter(rep IReporter) Comparer
	ObjectsEqual(x, y interface{}) bool
	ObjectsDiff(x, y interface{}) string
}

type objectsComparer struct {
	opts []cmp.Option
}

// NewObjectsComparer - конструктор интерфейса сравнения объектов
func NewObjectsComparer(opts ...cmp.Option) Comparer {
	return &objectsComparer{
		opts: opts,
	}
}

// WithIgnoreUnexportedOf - рекурсивно обходит переданные структуры и добавляет
// по ним и их полям структурного типа фильтр игнорирования приватных полей
func (oc *objectsComparer) WithIgnoreUnexportedOf(i ...interface{}) Comparer {
	oc.opts = append(oc.opts, BuildIgnoreUnexported(i...))

	return oc
}

// WithIgnoreEmptySlices - добавляет фильтр делающий тождественным пустой массив типа
// и пустой указатель на массив типа
func (oc *objectsComparer) WithIgnoreEmptySlices() Comparer {
	oc.opts = append(oc.opts, SlicesCompareOption())

	return oc
}

// WithSortSlices - применять сортировку срезов перед сравнением
func (oc *objectsComparer) WithSortSlices(lessFunc LessSliсes) Comparer {
	oc.opts = append(oc.opts, SortSlicesOption(lessFunc))

	return oc
}

// WithAllowUnexportedOf - рекурсивно обходит переданные структуры и добавляет
// по ним и их полям структурного типа фильтр разрешающий сравнивать приватные поля
func (oc *objectsComparer) WithAllowUnexportedOf(i ...interface{}) Comparer {
	oc.opts = append(oc.opts, BuildAllowUnexported(i...))

	return oc
}

// WithIgnoreFields - добавляет поля для игнорирования при сравнении для указанного типа
func (oc *objectsComparer) WithIgnoreFields(typ interface{}, names ...string) Comparer {
	oc.opts = append(oc.opts, cmpopts.IgnoreFields(typ, names...))

	return oc
}

// WithCustomReporter - устанавливает кастомный репортер `cmp`
func (oc *objectsComparer) WithCustomReporter(rep IReporter) Comparer {
	oc.opts = append(oc.opts, cmp.Reporter(rep))

	return oc
}

// ObjectsEqual - сравнивает две произвольных структуры, либо указателя на структуры,
// одного типа c игнорированием значений приватных полей
func (oc *objectsComparer) ObjectsEqual(x, y interface{}) bool {
	return objectsEqual(x, y, oc.opts...)
}

// ObjectsDiff - выводит diff двух произвольных структур одного типа
func (oc *objectsComparer) ObjectsDiff(x, y interface{}) string {
	if objectsEqual(x, y, oc.opts...) {
		return ""
	}

	return cmp.Diff(x, y, oc.opts...)
}

func objectsEqual(x, y interface{}, opts ...cmp.Option) bool {
	vx, vy := reflect.ValueOf(x), reflect.ValueOf(y)

	if !vx.CanInterface() || !vy.CanInterface() {
		return false
	}

	if vx.Type() != vy.Type() {
		return false
	}

	return cmp.Equal(x, y, opts...)
}
