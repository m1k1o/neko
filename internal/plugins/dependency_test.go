package plugins

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/demodesk/neko/pkg/types"
)

func Test_deps_addPlugin(t *testing.T) {
	type args struct {
		p []types.Plugin
	}
	tests := []struct {
		name     string
		args     args
		want     map[string]*dependency
		skipRun  bool
		wantErr1 bool
		wantErr2 bool
	}{
		{
			name: "three plugins - no dependencies",
			args: args{
				p: []types.Plugin{
					&dummyPlugin{name: "first"},
					&dummyPlugin{name: "second"},
					&dummyPlugin{name: "third"},
				},
			},
			want: map[string]*dependency{
				"first": {
					plugin:    &dummyPlugin{name: "first", idx: 0},
					invoked:   true,
					dependsOn: nil,
				},
				"second": {
					plugin:    &dummyPlugin{name: "second", idx: 0},
					invoked:   true,
					dependsOn: nil,
				},
				"third": {
					plugin:    &dummyPlugin{name: "third", idx: 0},
					invoked:   true,
					dependsOn: nil,
				},
			},
		}, {
			name: "three plugins - one dependency",
			args: args{
				p: []types.Plugin{
					&dummyPlugin{name: "third", dep: []string{"second"}},
					&dummyPlugin{name: "first"},
					&dummyPlugin{name: "second"},
				},
			},
			want: map[string]*dependency{
				"first": {
					plugin:    &dummyPlugin{name: "first", idx: 0},
					invoked:   true,
					dependsOn: nil,
				},
				"second": {
					plugin:    &dummyPlugin{name: "second", idx: 0},
					invoked:   true,
					dependsOn: nil,
				},
				"third": {
					plugin:  &dummyPlugin{name: "third", dep: []string{"second"}, idx: 1},
					invoked: true,
					dependsOn: []*dependency{
						{
							plugin:    &dummyPlugin{name: "second", idx: 0},
							invoked:   true,
							dependsOn: nil,
						},
					},
				},
			},
		}, {
			name: "three plugins - one double dependency",
			args: args{
				p: []types.Plugin{
					&dummyPlugin{name: "third", dep: []string{"first", "second"}},
					&dummyPlugin{name: "first"},
					&dummyPlugin{name: "second"},
				},
			},
			want: map[string]*dependency{
				"first": {
					plugin:    &dummyPlugin{name: "first", idx: 0},
					invoked:   true,
					dependsOn: nil,
				},
				"second": {
					plugin:    &dummyPlugin{name: "second", idx: 0},
					invoked:   true,
					dependsOn: nil,
				},
				"third": {
					plugin:  &dummyPlugin{name: "third", dep: []string{"first", "second"}, idx: 1},
					invoked: true,
					dependsOn: []*dependency{
						{
							plugin:    &dummyPlugin{name: "first", idx: 0},
							invoked:   true,
							dependsOn: nil,
						},
						{
							plugin:    &dummyPlugin{name: "second", idx: 0},
							invoked:   true,
							dependsOn: nil,
						},
					},
				},
			},
		}, {
			name: "three plugins - two dependencies",
			args: args{
				p: []types.Plugin{
					&dummyPlugin{name: "third", dep: []string{"first"}},
					&dummyPlugin{name: "first"},
					&dummyPlugin{name: "second", dep: []string{"first"}},
				},
			},
			want: map[string]*dependency{
				"first": {
					plugin:    &dummyPlugin{name: "first"},
					invoked:   false,
					dependsOn: nil,
				},
				"third": {
					plugin:  &dummyPlugin{name: "third", dep: []string{"first"}},
					invoked: false,
					dependsOn: []*dependency{
						{
							plugin:    &dummyPlugin{name: "first"},
							invoked:   false,
							dependsOn: nil,
						},
					},
				},
				"second": {
					plugin:  &dummyPlugin{name: "second", dep: []string{"first"}},
					invoked: false,
					dependsOn: []*dependency{
						{
							plugin:    &dummyPlugin{name: "first"},
							invoked:   false,
							dependsOn: nil,
						},
					},
				},
			},
			skipRun: true,
		}, {
			name: "three plugins - three dependencies",
			args: args{
				p: []types.Plugin{
					&dummyPlugin{name: "third", dep: []string{"second"}},
					&dummyPlugin{name: "first"},
					&dummyPlugin{name: "second", dep: []string{"first"}},
				},
			},
			want: map[string]*dependency{
				"first": {
					plugin:    &dummyPlugin{name: "first", idx: 0},
					invoked:   true,
					dependsOn: nil,
				},
				"second": {
					plugin:  &dummyPlugin{name: "second", dep: []string{"first"}, idx: 1},
					invoked: true,
					dependsOn: []*dependency{
						{
							plugin:    &dummyPlugin{name: "first", idx: 0},
							invoked:   true,
							dependsOn: nil,
						},
					},
				},
				"third": {
					plugin:  &dummyPlugin{name: "third", dep: []string{"second"}, idx: 2},
					invoked: true,
					dependsOn: []*dependency{
						{
							plugin:  &dummyPlugin{name: "second", dep: []string{"first"}, idx: 1},
							invoked: true,
							dependsOn: []*dependency{
								{
									plugin:    &dummyPlugin{name: "first", idx: 0},
									invoked:   true,
									dependsOn: nil,
								},
							},
						},
					},
				},
			},
		}, {
			name: "four plugins - added in reverse order, with dependencies",
			args: args{
				p: []types.Plugin{
					&dummyPlugin{name: "forth", dep: []string{"third"}},
					&dummyPlugin{name: "third", dep: []string{"second"}},
					&dummyPlugin{name: "second", dep: []string{"first"}},
					&dummyPlugin{name: "first"},
				},
			},
			want: map[string]*dependency{
				"first": {
					plugin:    &dummyPlugin{name: "first", idx: 0},
					invoked:   false,
					dependsOn: nil,
				},
				"second": {
					plugin:  &dummyPlugin{name: "second", dep: []string{"first"}, idx: 0},
					invoked: false,
					dependsOn: []*dependency{
						{
							plugin:    &dummyPlugin{name: "first", idx: 0},
							invoked:   false,
							dependsOn: nil,
						},
					},
				},
				"third": {
					plugin:  &dummyPlugin{name: "third", dep: []string{"second"}, idx: 0},
					invoked: false,
					dependsOn: []*dependency{
						{
							plugin:  &dummyPlugin{name: "second", dep: []string{"first"}, idx: 0},
							invoked: false,
							dependsOn: []*dependency{
								{
									plugin:    &dummyPlugin{name: "first", idx: 0},
									invoked:   false,
									dependsOn: nil,
								},
							},
						},
					},
				},
				"forth": {
					plugin:  &dummyPlugin{name: "forth", dep: []string{"third"}, idx: 0},
					invoked: false,
					dependsOn: []*dependency{
						{
							plugin:  &dummyPlugin{name: "third", dep: []string{"second"}, idx: 0},
							invoked: false,
							dependsOn: []*dependency{
								{
									plugin:  &dummyPlugin{name: "second", dep: []string{"first"}, idx: 0},
									invoked: false,
									dependsOn: []*dependency{
										{
											plugin:    &dummyPlugin{name: "first", idx: 0},
											invoked:   false,
											dependsOn: nil,
										},
									},
								},
							},
						},
					},
				},
			},
			skipRun: true,
		}, {
			name: "four plugins - two double dependencies",
			args: args{
				p: []types.Plugin{
					&dummyPlugin{name: "forth", dep: []string{"first", "third"}},
					&dummyPlugin{name: "third", dep: []string{"first", "second"}},
					&dummyPlugin{name: "second"},
					&dummyPlugin{name: "first"},
				},
			},
			want: map[string]*dependency{
				"first": {
					plugin:    &dummyPlugin{name: "first", idx: 0},
					invoked:   true,
					dependsOn: nil,
				},
				"second": {
					plugin:    &dummyPlugin{name: "second", idx: 0},
					invoked:   true,
					dependsOn: nil,
				},
				"third": {
					plugin:  &dummyPlugin{name: "third", dep: []string{"first", "second"}, idx: 1},
					invoked: true,
					dependsOn: []*dependency{
						{
							plugin:    &dummyPlugin{name: "first", idx: 0},
							invoked:   true,
							dependsOn: nil,
						},
						{
							plugin:    &dummyPlugin{name: "second", idx: 0},
							invoked:   true,
							dependsOn: nil,
						},
					},
				},
				"forth": {
					plugin:  &dummyPlugin{name: "forth", dep: []string{"first", "third"}, idx: 2},
					invoked: true,
					dependsOn: []*dependency{
						{
							plugin:    &dummyPlugin{name: "first", idx: 0},
							invoked:   true,
							dependsOn: nil,
						},
						{
							plugin:  &dummyPlugin{name: "third", dep: []string{"first", "second"}, idx: 1},
							invoked: true,
							dependsOn: []*dependency{
								{
									plugin:    &dummyPlugin{name: "first", idx: 0},
									invoked:   true,
									dependsOn: nil,
								},
								{
									plugin:    &dummyPlugin{name: "second", idx: 0},
									invoked:   true,
									dependsOn: nil,
								},
							},
						},
					},
				},
			},
		}, {
			// So, when we have plugin A in the list and want to add plugin C we can't determine the proper order without
			// resolving their direct dependiencies first:
			//
			// Can be C->D->A->B if D depends on A
			//
			// So to do it properly I would imagine tht we need to resolve all direct dependiencies first and build multiple lists:
			//
			// i.e. A->B->C D F->G
			//
			// and then join these lists in any order.
			name: "add indirect dependency CDAB",
			args: args{
				p: []types.Plugin{
					&dummyPlugin{name: "A", dep: []string{"B"}},
					&dummyPlugin{name: "C", dep: []string{"D"}},
					&dummyPlugin{name: "B"},
					&dummyPlugin{name: "D", dep: []string{"A"}},
				},
			},
			want: map[string]*dependency{
				"B": {
					plugin:    &dummyPlugin{name: "B", idx: 0},
					invoked:   true,
					dependsOn: nil,
				},
				"A": {
					plugin:  &dummyPlugin{name: "A", dep: []string{"B"}, idx: 1},
					invoked: true,
					dependsOn: []*dependency{
						{
							plugin:    &dummyPlugin{name: "B", idx: 0},
							invoked:   true,
							dependsOn: nil,
						},
					},
				},
				"D": {
					plugin:  &dummyPlugin{name: "D", dep: []string{"A"}, idx: 2},
					invoked: true,
					dependsOn: []*dependency{
						{
							plugin:  &dummyPlugin{name: "A", dep: []string{"B"}, idx: 1},
							invoked: true,
							dependsOn: []*dependency{
								{
									plugin:    &dummyPlugin{name: "B", idx: 0},
									invoked:   true,
									dependsOn: nil,
								},
							},
						},
					},
				},
				"C": {
					plugin:  &dummyPlugin{name: "C", dep: []string{"D"}, idx: 3},
					invoked: true,
					dependsOn: []*dependency{
						{
							plugin:  &dummyPlugin{name: "D", dep: []string{"A"}, idx: 2},
							invoked: true,
							dependsOn: []*dependency{
								{
									plugin:  &dummyPlugin{name: "A", dep: []string{"B"}, idx: 1},
									invoked: true,
									dependsOn: []*dependency{
										{
											plugin:    &dummyPlugin{name: "B", idx: 0},
											invoked:   true,
											dependsOn: nil,
										},
									},
								},
							},
						},
					},
				},
			},
		}, {
			// So, when we have plugin A in the list and want to add plugin C we can't determine the proper order without
			// resolving their direct dependiencies first:
			//
			// Can be A->B->C->D (in this test) if B depends on C
			//
			// So to do it properly I would imagine tht we need to resolve all direct dependiencies first and build multiple lists:
			//
			// i.e. A->B->C D F->G
			//
			// and then join these lists in any order.
			name: "add indirect dependency ABCD",
			args: args{
				p: []types.Plugin{
					&dummyPlugin{name: "C", dep: []string{"D"}},
					&dummyPlugin{name: "D"},
					&dummyPlugin{name: "B", dep: []string{"C"}},
					&dummyPlugin{name: "A", dep: []string{"B"}},
				},
			},
			want: map[string]*dependency{
				"D": {
					plugin:    &dummyPlugin{name: "D", idx: 0},
					invoked:   true,
					dependsOn: nil,
				},
				"C": {
					plugin:  &dummyPlugin{name: "C", dep: []string{"D"}, idx: 1},
					invoked: true,
					dependsOn: []*dependency{
						{
							plugin:    &dummyPlugin{name: "D", idx: 0},
							invoked:   true,
							dependsOn: nil,
						},
					},
				},
				"B": {
					plugin:  &dummyPlugin{name: "B", dep: []string{"C"}, idx: 2},
					invoked: true,
					dependsOn: []*dependency{
						{
							plugin:  &dummyPlugin{name: "C", dep: []string{"D"}, idx: 1},
							invoked: true,
							dependsOn: []*dependency{
								{
									plugin:    &dummyPlugin{name: "D", idx: 0},
									invoked:   true,
									dependsOn: nil,
								},
							},
						},
					},
				},
				"A": {
					plugin:  &dummyPlugin{name: "A", dep: []string{"B"}, idx: 3},
					invoked: true,
					dependsOn: []*dependency{
						{
							plugin:  &dummyPlugin{name: "B", dep: []string{"C"}, idx: 2},
							invoked: true,
							dependsOn: []*dependency{
								{
									plugin:  &dummyPlugin{name: "C", dep: []string{"D"}, idx: 1},
									invoked: true,
									dependsOn: []*dependency{
										{
											plugin:    &dummyPlugin{name: "D", idx: 0},
											invoked:   true,
											dependsOn: nil,
										},
									},
								},
							},
						},
					},
				},
			},
		}, {
			name: "add duplicate plugin",
			args: args{
				p: []types.Plugin{
					&dummyPlugin{name: "first"},
					&dummyPlugin{name: "first"},
				},
			},
			want: map[string]*dependency{
				"first": {plugin: &dummyPlugin{name: "first", idx: 0}, invoked: true},
			},
			wantErr1: true,
		}, {
			name: "cyclical dependency",
			args: args{
				p: []types.Plugin{
					&dummyPlugin{name: "first", dep: []string{"second"}},
					&dummyPlugin{name: "second", dep: []string{"first"}},
				},
			},
			want: map[string]*dependency{
				"first": {
					plugin:  &dummyPlugin{name: "first", dep: []string{"second"}, idx: 1},
					invoked: true,
				},
			},
			wantErr1: true,
		}, {
			name: "four plugins - cyclical transitive dependencies in reverse order",
			args: args{
				p: []types.Plugin{
					&dummyPlugin{name: "forth", dep: []string{"third"}},
					&dummyPlugin{name: "third", dep: []string{"second"}},
					&dummyPlugin{name: "second", dep: []string{"first"}},
					&dummyPlugin{name: "first", dep: []string{"forth"}},
				},
			},
			want: map[string]*dependency{
				"second": {
					plugin:  &dummyPlugin{name: "second", dep: []string{"first"}, idx: 0},
					invoked: false,
					dependsOn: []*dependency{
						{
							plugin:    &dummyPlugin{name: "first", dep: []string{"forth"}, idx: 0},
							invoked:   false,
							dependsOn: nil,
						},
					},
				},
				"third": {
					plugin:  &dummyPlugin{name: "third", dep: []string{"second"}, idx: 0},
					invoked: false,
					dependsOn: []*dependency{
						{
							plugin:  &dummyPlugin{name: "second", dep: []string{"first"}, idx: 0},
							invoked: false,
							dependsOn: []*dependency{
								{
									plugin:    &dummyPlugin{name: "first", dep: []string{"forth"}, idx: 0},
									invoked:   false,
									dependsOn: nil,
								},
							},
						},
					},
				},
				"forth": {
					plugin:    &dummyPlugin{name: "forth", dep: []string{"third"}, idx: 0},
					invoked:   false,
					dependsOn: nil,
				},
			},
			wantErr1: true,
			skipRun:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &dependiencies{deps: make(map[string]*dependency)}

			var (
				err     error
				counter int
			)
			for _, p := range tt.args.p {
				if !tt.skipRun {
					p.(*dummyPlugin).counter = &counter
				}
				if err = d.addPlugin(p); err != nil {
					break
				}
			}
			if err != nil != tt.wantErr1 {
				t.Errorf("dependiencies.addPlugin() error = %v, wantErr1 %v", err, tt.wantErr1)
				return
			}

			if !tt.skipRun {
				if err := d.start(types.PluginManagers{}); (err != nil) != tt.wantErr2 {
					t.Errorf("dependiencies.start() error = %v, wantErr1 %v", err, tt.wantErr2)
				}
			}

			assert.Equal(t, tt.want, d.deps)
		})
	}
}

type dummyPlugin struct {
	name    string
	dep     []string
	idx     int
	counter *int
}

func (d dummyPlugin) Name() string {
	return d.name
}

func (d dummyPlugin) DependsOn() []string {
	return d.dep
}

func (d dummyPlugin) Config() types.PluginConfig {
	return nil
}

func (d *dummyPlugin) Start(types.PluginManagers) error {
	if len(d.dep) > 0 {
		*d.counter++
		d.idx = *d.counter
	}
	d.counter = nil
	return nil
}

func (d dummyPlugin) Shutdown() error {
	return nil
}
