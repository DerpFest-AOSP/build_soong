// Copyright 2015 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package android

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/blueprint"
)

type mutatorTestModule struct {
	ModuleBase
	props struct {
		Deps_missing_deps    []string
		Mutator_missing_deps []string
	}

	missingDeps []string
}

func mutatorTestModuleFactory() Module {
	module := &mutatorTestModule{}
	module.AddProperties(&module.props)
	InitAndroidModule(module)
	return module
}

func (m *mutatorTestModule) GenerateAndroidBuildActions(ctx ModuleContext) {
	ctx.Build(pctx, BuildParams{
		Rule:   Touch,
		Output: PathForModuleOut(ctx, "output"),
	})

	m.missingDeps = ctx.GetMissingDependencies()
}

func (m *mutatorTestModule) DepsMutator(ctx BottomUpMutatorContext) {
	ctx.AddDependency(ctx.Module(), nil, m.props.Deps_missing_deps...)
}

func addMissingDependenciesMutator(ctx TopDownMutatorContext) {
	ctx.AddMissingDependencies(ctx.Module().(*mutatorTestModule).props.Mutator_missing_deps)
}

func TestMutatorAddMissingDependencies(t *testing.T) {
	bp := `
		test {
			name: "foo",
			deps_missing_deps: ["regular_missing_dep"],
			mutator_missing_deps: ["added_missing_dep"],
		}
	`

	result := GroupFixturePreparers(
		PrepareForTestWithAllowMissingDependencies,
		FixtureRegisterWithContext(func(ctx RegistrationContext) {
			ctx.RegisterModuleType("test", mutatorTestModuleFactory)
			ctx.PreDepsMutators(func(ctx RegisterMutatorsContext) {
				ctx.TopDown("add_missing_dependencies", addMissingDependenciesMutator)
			})
		}),
		FixtureWithRootAndroidBp(bp),
	).RunTest(t)

	foo := result.ModuleForTests("foo", "").Module().(*mutatorTestModule)

	AssertDeepEquals(t, "foo missing deps", []string{"added_missing_dep", "regular_missing_dep"}, foo.missingDeps)
}

type testTransitionMutator struct {
	split              func(ctx BaseModuleContext) []string
	outgoingTransition func(ctx OutgoingTransitionContext, sourceVariation string) string
	incomingTransition func(ctx IncomingTransitionContext, incomingVariation string) string
	mutate             func(ctx BottomUpMutatorContext, variation string)
}

func (t *testTransitionMutator) Split(ctx BaseModuleContext) []string {
	if t.split != nil {
		return t.split(ctx)
	}
	return []string{""}
}

func (t *testTransitionMutator) OutgoingTransition(ctx OutgoingTransitionContext, sourceVariation string) string {
	if t.outgoingTransition != nil {
		return t.outgoingTransition(ctx, sourceVariation)
	}
	return sourceVariation
}

func (t *testTransitionMutator) IncomingTransition(ctx IncomingTransitionContext, incomingVariation string) string {
	if t.incomingTransition != nil {
		return t.incomingTransition(ctx, incomingVariation)
	}
	return incomingVariation
}

func (t *testTransitionMutator) Mutate(ctx BottomUpMutatorContext, variation string) {
	if t.mutate != nil {
		t.mutate(ctx, variation)
	}
}

func TestModuleString(t *testing.T) {
	bp := `
		test {
			name: "foo",
		}
	`

	var moduleStrings []string

	GroupFixturePreparers(
		FixtureRegisterWithContext(func(ctx RegistrationContext) {

			ctx.PreArchMutators(func(ctx RegisterMutatorsContext) {
				ctx.Transition("pre_arch", &testTransitionMutator{
					split: func(ctx BaseModuleContext) []string {
						moduleStrings = append(moduleStrings, ctx.Module().String())
						return []string{"a", "b"}
					},
				})
				ctx.TopDown("rename_top_down", func(ctx TopDownMutatorContext) {
					moduleStrings = append(moduleStrings, ctx.Module().String())
					ctx.Rename(ctx.Module().base().Name() + "_renamed1")
				})
			})

			ctx.PreDepsMutators(func(ctx RegisterMutatorsContext) {
				ctx.Transition("pre_deps", &testTransitionMutator{
					split: func(ctx BaseModuleContext) []string {
						moduleStrings = append(moduleStrings, ctx.Module().String())
						return []string{"c", "d"}
					},
				})
			})

			ctx.PostDepsMutators(func(ctx RegisterMutatorsContext) {
				ctx.Transition("post_deps", &testTransitionMutator{
					split: func(ctx BaseModuleContext) []string {
						moduleStrings = append(moduleStrings, ctx.Module().String())
						return []string{"e", "f"}
					},
					outgoingTransition: func(ctx OutgoingTransitionContext, sourceVariation string) string {
						return ""
					},
				})
				ctx.BottomUp("rename_bottom_up", func(ctx BottomUpMutatorContext) {
					moduleStrings = append(moduleStrings, ctx.Module().String())
					ctx.Rename(ctx.Module().base().Name() + "_renamed2")
				})
				ctx.BottomUp("final", func(ctx BottomUpMutatorContext) {
					moduleStrings = append(moduleStrings, ctx.Module().String())
				})
			})

			ctx.RegisterModuleType("test", mutatorTestModuleFactory)
		}),
		FixtureWithRootAndroidBp(bp),
	).RunTest(t)

	want := []string{
		// Initial name.
		"foo{}",

		// After pre_arch (reversed because rename_top_down is TopDown so it visits in reverse order).
		"foo{pre_arch:b}",
		"foo{pre_arch:a}",

		// After rename_top_down (reversed because pre_deps TransitionMutator.Split is TopDown).
		"foo_renamed1{pre_arch:b}",
		"foo_renamed1{pre_arch:a}",

		// After pre_deps (reversed because post_deps TransitionMutator.Split is TopDown).
		"foo_renamed1{pre_arch:b,pre_deps:d}",
		"foo_renamed1{pre_arch:b,pre_deps:c}",
		"foo_renamed1{pre_arch:a,pre_deps:d}",
		"foo_renamed1{pre_arch:a,pre_deps:c}",

		// After post_deps.
		"foo_renamed1{pre_arch:a,pre_deps:c,post_deps:e}",
		"foo_renamed1{pre_arch:a,pre_deps:c,post_deps:f}",
		"foo_renamed1{pre_arch:a,pre_deps:d,post_deps:e}",
		"foo_renamed1{pre_arch:a,pre_deps:d,post_deps:f}",
		"foo_renamed1{pre_arch:b,pre_deps:c,post_deps:e}",
		"foo_renamed1{pre_arch:b,pre_deps:c,post_deps:f}",
		"foo_renamed1{pre_arch:b,pre_deps:d,post_deps:e}",
		"foo_renamed1{pre_arch:b,pre_deps:d,post_deps:f}",

		// After rename_bottom_up.
		"foo_renamed2{pre_arch:a,pre_deps:c,post_deps:e}",
		"foo_renamed2{pre_arch:a,pre_deps:c,post_deps:f}",
		"foo_renamed2{pre_arch:a,pre_deps:d,post_deps:e}",
		"foo_renamed2{pre_arch:a,pre_deps:d,post_deps:f}",
		"foo_renamed2{pre_arch:b,pre_deps:c,post_deps:e}",
		"foo_renamed2{pre_arch:b,pre_deps:c,post_deps:f}",
		"foo_renamed2{pre_arch:b,pre_deps:d,post_deps:e}",
		"foo_renamed2{pre_arch:b,pre_deps:d,post_deps:f}",
	}

	AssertDeepEquals(t, "module String() values", want, moduleStrings)
}

func TestFinalDepsPhase(t *testing.T) {
	bp := `
		test {
			name: "common_dep_1",
		}
		test {
			name: "common_dep_2",
		}
		test {
			name: "foo",
		}
	`

	finalGot := map[string]int{}

	GroupFixturePreparers(
		FixtureRegisterWithContext(func(ctx RegistrationContext) {
			dep1Tag := struct {
				blueprint.BaseDependencyTag
			}{}
			dep2Tag := struct {
				blueprint.BaseDependencyTag
			}{}

			ctx.PostDepsMutators(func(ctx RegisterMutatorsContext) {
				ctx.BottomUp("far_deps_1", func(ctx BottomUpMutatorContext) {
					if !strings.HasPrefix(ctx.ModuleName(), "common_dep") {
						ctx.AddFarVariationDependencies([]blueprint.Variation{}, dep1Tag, "common_dep_1")
					}
				})
				ctx.Transition("variant", &testTransitionMutator{
					split: func(ctx BaseModuleContext) []string {
						return []string{"a", "b"}
					},
				})
			})

			ctx.FinalDepsMutators(func(ctx RegisterMutatorsContext) {
				ctx.BottomUp("far_deps_2", func(ctx BottomUpMutatorContext) {
					if !strings.HasPrefix(ctx.ModuleName(), "common_dep") {
						ctx.AddFarVariationDependencies([]blueprint.Variation{}, dep2Tag, "common_dep_2")
					}
				})
				ctx.BottomUp("final", func(ctx BottomUpMutatorContext) {
					finalGot[ctx.Module().String()] += 1
					ctx.VisitDirectDeps(func(mod Module) {
						finalGot[fmt.Sprintf("%s -> %s", ctx.Module().String(), mod)] += 1
					})
				})
			})

			ctx.RegisterModuleType("test", mutatorTestModuleFactory)
		}),
		FixtureWithRootAndroidBp(bp),
	).RunTest(t)

	finalWant := map[string]int{
		"common_dep_1{variant:a}":                   1,
		"common_dep_1{variant:b}":                   1,
		"common_dep_2{variant:a}":                   1,
		"common_dep_2{variant:b}":                   1,
		"foo{variant:a}":                            1,
		"foo{variant:a} -> common_dep_1{variant:a}": 1,
		"foo{variant:a} -> common_dep_2{variant:a}": 1,
		"foo{variant:b}":                            1,
		"foo{variant:b} -> common_dep_1{variant:b}": 1,
		"foo{variant:b} -> common_dep_2{variant:a}": 1,
	}

	AssertDeepEquals(t, "final", finalWant, finalGot)
}

func TestNoCreateVariationsInFinalDeps(t *testing.T) {
	GroupFixturePreparers(
		FixtureRegisterWithContext(func(ctx RegistrationContext) {
			ctx.FinalDepsMutators(func(ctx RegisterMutatorsContext) {
				ctx.Transition("vars", &testTransitionMutator{
					split: func(ctx BaseModuleContext) []string {
						return []string{"a", "b"}
					},
				})
			})

			ctx.RegisterModuleType("test", mutatorTestModuleFactory)
		}),
		FixtureWithRootAndroidBp(`test {name: "foo"}`),
	).
		ExtendWithErrorHandler(FixtureExpectsOneErrorPattern("not allowed in FinalDepsMutators")).
		RunTest(t)
}
