/*
 *	Copyright 2024 Jan Pfeifer
 *
 *	Licensed under the Apache License, Version 2.0 (the "License");
 *	you may not use this file except in compliance with the License.
 *	You may obtain a copy of the License at
 *
 *	http://www.apache.org/licenses/LICENSE-2.0
 *
 *	Unless required by applicable law or agreed to in writing, software
 *	distributed under the License is distributed on an "AS IS" BASIS,
 *	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *	See the License for the specific language governing permissions and
 *	limitations under the License.
 */

package graph_test

import (
	"fmt"
	. "github.com/gomlx/gomlx/graph"
	"github.com/gomlx/gomlx/graph/graphtest"
	"github.com/gomlx/gomlx/types/shapes"
	"github.com/gomlx/gomlx/types/tensors"
	"github.com/gomlx/gopjrt/dtypes"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/rand"
	"slices"
	"testing"
)

func TestIndicesForShape(t *testing.T) {
	backend := graphtest.BuildTestBackend()
	g := NewGraph(backend, t.Name())
	shape := MakeShape(F64, 2, 3, 4)
	numbers := IndicesForShape(g, shape)
	g.Compile(numbers)
	got := g.Run()[0]
	fmt.Printf("\tIndicesForShape(%s)=%v\n", shape, got)
	want := [][]int64{{0, 0, 0}, {0, 0, 1}, {0, 0, 2}, {0, 0, 3}, {0, 1, 0}, {0, 1, 1}, {0, 1, 2}, {0, 1, 3}, {0, 2, 0}, {0, 2, 1}, {0, 2, 2}, {0, 2, 3}, {1, 0, 0}, {1, 0, 1}, {1, 0, 2}, {1, 0, 3}, {1, 1, 0}, {1, 1, 1}, {1, 1, 2}, {1, 1, 3}, {1, 2, 0}, {1, 2, 1}, {1, 2, 2}, {1, 2, 3}}
	require.Equalf(t, want, got.Value(), "IndicesForShape(%s): want %v, got %v", shape, want, got)
}

func TestGather(t *testing.T) {
	backend := graphtest.BuildTestBackend()
	{ // Trivial scalar gather.
		fmt.Println("\tGather(): trivial scalar gather.")
		g := NewGraph(backend, t.Name())
		// numbers=(Float64)[5 3]: [[0 1 2] [3 4 5] [6 7 8] [9 10 11] [12 13 14]]
		numbers := IotaFull(g, MakeShape(F64, 5, 3))
		indices := Const(g, 1)
		gather := Gather(numbers, indices)
		g.Compile(gather)
		got := g.Run()[0]
		fmt.Printf("\t\tGather=%v\n", got)
		want := []float64{3, 4, 5}
		require.Equalf(t, want, got.Value(), "Gather: want %v, got %v", want, got)
	}

	{ // Simple leading indices dimension.
		fmt.Println("\tGather(): simple leading indices dimension.")
		g := NewGraph(backend, "Gather(): simple leading indices dimension.")
		// numbers=(Float64)[5 3]: [[0 1 2] [3 4 5] [6 7 8] [9 10 11] [12 13 14]]
		numbers := IotaFull(g, MakeShape(F64, 5, 3))
		indices := Const(g, [][]int{{2}, {0}})
		gather := Gather(numbers, indices)
		g.Compile(gather)
		got := g.Run()[0]
		fmt.Printf("\t\tGather=%v\n", got)
		want := [][]float64{{6, 7, 8}, {0, 1, 2}}
		require.Equalf(t, want, got.Value(), "Gather: want %v, got %v", want, got)
	}

	{ // With 2D leading indices dimension.
		fmt.Println("\tGather(): with 2D leading indices dimension.")
		g := NewGraph(backend, "Gather(): with 2D leading indices dimension.")
		// numbers=(Float64)[5 3]: [[0 1 2] [3 4 5] [6 7 8] [9 10 11] [12 13 14]]
		numbers := IotaFull(g, MakeShape(F64, 5, 3))
		indices := Const(g, [][][]int{{{2}, {0}}, {{2}, {1}}})
		gather := Gather(numbers, indices)
		g.Compile(gather)
		got := g.Run()[0]
		fmt.Printf("\t\tGather=%v\n", got)
		want := [][][]float64{{{6, 7, 8}, {0, 1, 2}}, {{6, 7, 8}, {3, 4, 5}}}
		require.Equalf(t, want, got.Value(), "Gather: want %v, got %v", want, got)
	}

	{ // With leading indices dimension, and 3D params tailing dimensions.
		fmt.Println("\tGather(): With leading indices dimension, and 2D params tailing dimensions.")
		g := NewGraph(backend, "Gather(): With leading indices dimension, and 2D params tailing dimensions.")
		// numbers=(Float64)[5 3]: [[0 1 2] [3 4 5] [6 7 8] [9 10 11] [12 13 14]]
		numbers := IotaFull(g, MakeShape(F64, 5, 2, 2))
		indices := Const(g, [][]int{{2}, {0}, {1}, {3}})
		gather := Gather(numbers, indices)
		g.Compile(gather)
		got := g.Run()[0]
		fmt.Printf("\t\tGather=%v\n", got)
		want := [][][]float64{{{8, 9}, {10, 11}}, {{0, 1}, {2, 3}}, {{4, 5}, {6, 7}}, {{12, 13}, {14, 15}}}
		require.Equalf(t, want, got.Value(), "Gather: want %v, got %v", want, got)
	}

}

func TestGatherSlices(t *testing.T) {
	testFuncOneInput(t, "GatherSlices(input, slicedAxes={1}, start={{0}, {1}, {0}}, sizes={1})",
		func(g *Graph) (input, output *Node) {
			input = IotaFull(g, shapes.Make(dtypes.Float32, 4, 5))
			start := Const(g, [][]int32{{0}, {1}, {0}}) // Slice from rows 0, 2 and 0 of each example in the batch.
			sizes := []int{1}                           // Take only one row per start.
			output = GatherSlices(input, []int{0}, start, sizes)
			return
		}, [][][]float32{{{0, 1, 2, 3, 4}}, {{5, 6, 7, 8, 9}}, {{0, 1, 2, 3, 4}}})

	testFuncOneInput(t, "GatherSlices(input, slicedAxes={0}, start={{0}, {1}}, sizes={2})",
		func(g *Graph) (input, output *Node) {
			input = IotaFull(g, shapes.Make(dtypes.Float32, 4, 3))
			start := Const(g, [][]int32{{0}, {1}}) // Slice from rows 0 and 1.
			sizes := []int{2}                      // Take two rows per start.
			output = GatherSlices(input, []int{0}, start, sizes)
			return
		}, [][][]float32{{{0, 1, 2}, {3, 4, 5}}, {{3, 4, 5}, {6, 7, 8}}})

	testFuncOneInput(t, "GatherSlices(input, slicedAxes={0,1}, start={1, 1}, sizes={2, 3})",
		func(g *Graph) (input, output *Node) {
			input = IotaFull(g, shapes.Make(dtypes.Float32, 4, 10))
			start := Const(g, []int32{1, 1}) // Slice in middle of matrix.
			sizes := []int{2, 3}             // Take a sub-matrix
			output = GatherSlices(input, []int{0, 1}, start, sizes)
			return
		}, [][]float32{{11, 12, 13}, {21, 22, 23}})
}

func TestScatter(t *testing.T) {
	backend := graphtest.BuildTestBackend()
	{ // Trivial scalar scatter.
		fmt.Println("\tScatter(): trivial scalar scatter.")
		g := NewGraph(backend, "Scatter(): trivial scalar scatter")
		// numbers=(Float64)[3]: [2 3 4]
		numbers := Add(IotaFull(g, MakeShape(F64, 3)), Const(g, float64(2)))
		indices := Const(g, 1)
		scatter := Scatter(indices, numbers, MakeShape(F64, 2, 3))
		g.Compile(scatter)
		got := g.Run()[0]
		fmt.Printf("\t\tscatter=%v\n", got)
		want := [][]float64{{0, 0, 0}, {2, 3, 4}}
		require.Equalf(t, want, got.Value(), "Scatter: want %v, got %v", want, got)
	}

	{ // Simple leading indices dimension.
		fmt.Println("\tScatterAdd(): leading indices dimension, and deeper slice dimension.")
		g := NewGraph(backend, "ScatterAdd(): leading indices dimension, and deeper slice dimension.")
		// numbers=(Float64)[5 3, 1]: [[[0] [1] [2]] [[3] [4] [5]]]
		numbers := IotaFull(g, MakeShape(F64, 2, 3, 1))
		indices := Const(g, [][]int{{2}, {0}})
		operand := Ones(g, MakeShape(F64, 3, 3, 1))
		scatter := ScatterAdd(operand, indices, numbers, false, true)
		g.Compile(scatter)
		got := g.Run()[0]
		fmt.Printf("\t\tscatter=%v\n", got)
		want := [][][]float64{{{4}, {5}, {6}}, {{1}, {1}, {1}}, {{1}, {2}, {3}}}
		require.Equalf(t, want, got.Value(), "Scatter: want %v, got %v", want, got)
	}
}

// BenchmarkScatter tests the various scatter combinations: sorted or unique and different dtypes.
// The auto-differentiation of a gather is a scatter: it is used in update of large embedding tables.
func BenchmarkScatter(b *testing.B) {
	backend := graphtest.BuildTestBackend()
	const (
		NumEntries          = 1_000_000
		EmbeddingSize       = 32
		BatchSize           = 100 // Number of indices to scatter
		ConsecutiveScatters = 100
	)
	indices := make([]int32, BatchSize)
	for ii := range indices {
		indices[ii] = int32(rand.Int31n(NumEntries - ConsecutiveScatters))
	}
	slices.Sort(indices)
	indicesT := tensors.FromValue(indices)
	rngStateT := RngStateFromSeed(42)

	for _, sorted := range []bool{true, false} {
		for _, unique := range []bool{true, false} {
			scatterExec := NewExec(backend, func(state, indices, values *Node) *Node {
				g := values.Graph()
				dtype := values.DType()
				zeros := Zeros(g, shapes.Make(dtype, NumEntries, EmbeddingSize))
				indices = ExpandAxes(indices, -1)
				parts := make([]*Node, ConsecutiveScatters)
				for ii := range parts {
					parts[ii] = ExpandAxes(ScatterAdd(zeros, AddScalar(indices, float64(ii)), values, sorted, unique), -1)
				}
				x := ReduceSum(Concatenate(parts, -1), -1)
				return Add(state, x)
			})
			for _, dtype := range []dtypes.DType{dtypes.Float64, dtypes.Float32, dtypes.Float16} { //
				// Create random values tensor shaped [BatchSize, EmbeddingSize] of the given dtype.
				results := NewExec(backend, func(rngState *Node) (state, value *Node) {
					_, state = RandomNormal(rngState, shapes.Make(dtype, NumEntries, EmbeddingSize))
					_, value = RandomNormal(rngState, shapes.Make(dtype, BatchSize, EmbeddingSize))
					return
				}).Call(rngStateT)
				stateT, valuesT := results[0], results[1]

				// Precompile graph for given inputNodes. It also makes sure the inputNodes are transferred to the accelerator.
				scatterExec.Call(stateT, indicesT, valuesT)[0].FinalizeAll()
				b.Run(fmt.Sprintf("sorted-%v_unique-%v_dtype-%s", sorted, unique, dtype), func(b *testing.B) {
					for _ = range b.N {
						results := scatterExec.Call(stateT, indicesT, valuesT)
						stateT.FinalizeAll()
						stateT = results[0]
					}
				})
			}
		}
	}
}
