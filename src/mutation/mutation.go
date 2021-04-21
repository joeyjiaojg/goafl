package mutation

type Mutator struct {
	fn func([]byte, func([]string, []byte) bool, []string)
}

var mutators []Mutator

func init() {
	mutators = append(mutators, Mutator{
		bit_flip_1,
	})
}

func Mutate(input []byte, cb func([]string, []byte) bool, args []string) []byte {
	out := input
	for i := range mutators {
		mutators[i].fn(input, cb, args)
	}
	return out
}

func flip_bit(input []byte, s int) []byte {
	out := input
	out[s>>3] ^= 128 >> (s & 7)
	return out
}

func bit_flip_1(input []byte, cb func([]string, []byte) bool, args []string) {
	stage_max := len(input) << 3
	for i := 0; i < stage_max; i++ {
		out := flip_bit(input, (i >> 3))
		cb(args, out)
	}
}
