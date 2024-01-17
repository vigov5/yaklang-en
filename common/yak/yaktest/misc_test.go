package yaktest

import "testing"

func TestMisc(t *testing.T) {
	cases := []YakTestCase{
		{
			Name: "Test codec mmh3",
			Src: `
dump(codec.MMH3Hash32("asdfasdfasdfasdf"));
dump(codec.MMH3Hash128("asdfasdfasdfasdf"));
dump(codec.MMH3Hash128x64("asdfasdfasdfasdf"))
`,
		},
		{Name: "Test x.Map", Src: `dump(x.Map([1,2,3,4], func(i){println(i);return "123123"}))`},
		{Name: "Test x.Sum", Src: `dump(x.Reduce([1,2,3,4], func(pre,after){println(pre);return pre+after},2))`},
		{Name: "Test x.Filter", Src: `dump(x.Filter([1,2,3,4], func(i){println(i);return i>2}))`},
		{Name: "Test x.Find", Src: `dump(x.Find([1,2,3,4], func(i){println(i);return i>2}))`},
		{Name: "Test x.Foreach", Src: `dump(x.Foreach([1,2,3,4], func(i){println(i)}))`},
		{Name: "Test x.ForeachRight", Src: `dump(x.ForeachRight([1,2,true,3,4], func(i){println(i)}))`},
		{Name: "Test x .Contains", Src: `assert(x.Contains([1,2,true,3,4], true))`},
		{Name: "Test x.Contains2", Src: `assert(x.Contains([1,2,true,3,4], 1))`},
		{Name: "Test x.Contains3", Src: `assert(!x.Contains([1,2,true,3,4], 7))`},
		{Name: "Test x.Find", Src: `assert(!x.Contains([1,2,true,3,4], var))`},
		{Name: "Test x.Contains5", Src: `assert(!x.Contains([1,2,true,3,4], nil))`},
		{Name: "Test x.IndexOf", Src: `assert(2 == x.IndexOf([1,2,true,3,4], true))`},
		{Name: "Test x .IndexOf1", Src: `assert(3 != x.IndexOf([1,2,true,3,4], true))`},
		{Name: "tests x.IndexOf2", Src: `assert(3 != x.IndexOf([1,2,true,3,4], 4))`},
		{Name: "Test x.IndexOf3", Src: `assert(4 == x.IndexOf([1,2,true,3,4], 4))`},
		{Name: "Test x.Difference", Src: `dump(x.Difference([1,2,true,3,4], [2,true,5]))`},
		{Name: "Test x .Subtract", Src: `dump(x.Subtract([1,2,true,3,4], [2,true,5]))`},
		{Name: "Test x.Equal", Src: `assert(x.Equal(x.Subtract([1,2,true,3,4], [2,false,5]), [1,true,3,4]))`},
		{Name: "Test x.Chunk", Src: `assert(x.Equal(x.Chunk([1,2,true,3,4], 2)[1], [true,3]))`},
		{Name: "Test x.RemoveRepeat", Src: `assert(x.Equal(x.RemoveRepeat([1,2,true,3,true,4]), [1,2,true,3,4]))`},
		{Name: "Test x.Tail2", Src: `println("-----------------------------");dump(x.Tail([1,2,true,3,4]))`},
		{Name: "Test x.Tail", Src: `assert(x.Equal(x.Tail([1,2,true,3,4]), [2,true,3,4]))`},
		{Name: "Test x.Head", Src: `assert(x.Head([1,2,true,3,4]) == 1)`},
		{Name: "Test x.Drop", Src: `assert(x.Equal(x.Drop([1,2,true,3,4], 2), [true,3,4]))`},
		{Name: "Test x.Values ", Src: `assert(x.Equal(x.Values({1:2,3:"true"}), [2, "true"]))`},
		{Name: "Test x.Keys", Src: `assert(x.Equal(x.Keys({1:2,3:"true"}), [1,3]))`},
		{Name: "Test x.Reverse", Src: `assert(x.Equal(x.Reverse([3,1,2,3,5]), [5,3,2,1,3]))`},
		{Name: "Test x.Sum", Src: `dump(x.Sum([3,1,2,3,5]))`},

		//{Name: "Test x.Foreach", Src: `dump(x.Find([1,2,true,3,4], func(i){println(i)}))`},
		//{Name: "Test x.Head", Src: `dump(x.ToMap([{"a": 123},{"b":1234455,"a":111}],"a"))`},
		//{Name: "Test x.FlatMap", Src: `dump(x.FlatMap({"123": "abc", "bcc": 123}, func(i,v){println(i);return "123123"}))`},
	}

	Run("mmh3 Usability test", t, cases...)
}
