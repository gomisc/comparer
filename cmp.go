package comparer

import (
	"reflect"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// LessSliсes функция сравнения срезов
type LessSliсes func(x, y interface{}) bool

// SlicesCompareOption - при сравнении `cmp` фильтрует пустые слайсы
// и делает тождественным nil (указатель на отсутствующий слайс) и пустой []SomeType{}
func SlicesCompareOption() cmp.Option {
	return cmp.FilterValues(
		func(x, y interface{}) bool {
			if x == nil || y == nil {
				return false
			}

			xv, yv := reflect.ValueOf(x), reflect.ValueOf(y)
			xt, t1 := xv.Type(), yv.Type()

			return xt.Kind() == reflect.Slice &&
				t1.Kind() == reflect.Slice &&
				xv.Len() == 0 &&
				yv.Len() == 0 &&
				(xv.IsNil() != yv.IsNil())
		},
		cmp.Transformer("nilSlice", func(in interface{}) interface{} {
			return reflect.Zero(reflect.TypeOf(in)).Interface()
		}),
	)
}

// SortSlicesOption опция сортировки массивов при сравнении
func SortSlicesOption(lessSlices LessSliсes) cmp.Option {
	return cmpopts.SortSlices(func(sx, sy interface{}) bool {
		return lessSlices(sx, sy)
	})
}

// BuildIgnoreUnexported - рекурсивно формирует `cmpopts.IgnoreUnexported`
func BuildIgnoreUnexported(i ...interface{}) cmp.Option {
	return cmpopts.IgnoreUnexported(getStructItemsRecursiveOf(i...)...)
}

// BuildAllowUnexported - рекурсивно формирует `cmp.AllowUnexported`
func BuildAllowUnexported(i ...interface{}) cmp.Option {
	return cmp.AllowUnexported(getStructItemsRecursiveOf(i...)...)
}

// рекурсивно обходит переданные в качестве параметров структуры,
// находит поля структурного типа с количеством полей >=1 и возвращает их список
func getStructItemsRecursiveOf(i ...interface{}) (items []interface{}) {
	filter := make(map[string]struct{})

	for i := range findMessageWithUnexportedFields(i...) {
		typeStr := reflect.TypeOf(i).String()
		if _, ok := filter[typeStr]; !ok {
			items = append(items, i)
			filter[typeStr] = struct{}{}
		}
	}

	return items
}

func findMessageWithUnexportedFields(obj ...interface{}) <-chan interface{} {
	mchan := make(chan interface{})
	// рекурсивный проход
	go func() {
		walkMessages(mchan, obj...)
		close(mchan)
	}()

	return mchan
}

// nolint
func walkMessages(ch chan interface{}, src ...interface{}) {
	for _, m := range src {
		var children []interface{}

		ch <- m

		v := reflect.ValueOf(m)
		t := v.Type()

		for i := 0; i < t.NumField(); i++ {
			f := v.Field(i)
			// для структур которые являются слайсами внутри указанных структур
			if f.Kind() == reflect.Slice {
				if f.Type().Elem().Kind() == reflect.Ptr {
					if f.Type().Elem().Elem().Kind() == reflect.Struct {
						fa := reflect.New(f.Type().Elem())
						f = reflect.Indirect(fa)
					}
				}
			}

			// если поле - это указатель на структуру
			if f.Kind() == reflect.Ptr && f.CanInterface() {
				fa := reflect.New(f.Type().Elem())
				fv := reflect.Indirect(fa)
				children = append(children, fv.Interface())
			}

			// если поле - это структура количеством полей больше 0
			if f.Kind() == reflect.Struct {
				if f.Type().NumField() > 0 {
					fa := reflect.New(f.Type())
					fv := reflect.Indirect(fa)
					children = append(children, fv.Interface())
				}
			}
		}

		walkMessages(ch, children...)
	}
}
