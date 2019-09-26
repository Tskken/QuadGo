package quadgo

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestSetMaxEntities(t *testing.T) {
	type fields struct {
		options *Options
	}
	type args struct {
		maxEntities uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   uint64
	}{
		{
			name: "basic set max entities",
			fields: fields{
				options: &Options{
					MaxEntities: 1,
					MaxDepth:    1,
				},
			},
			args: args{
				maxEntities: 5,
			},
			want: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opFunc := SetMaxEntities(tt.args.maxEntities)
			opFunc(tt.fields.options)
			if tt.fields.options.MaxEntities != tt.want {
				t.Errorf("quadgo.SetMaxEntities() = %v, want %v", tt.fields.options.MaxEntities, tt.want)
			}
		})
	}
}

func TestSetMaxDepth(t *testing.T) {
	type fields struct {
		options *Options
	}
	type args struct {
		maxDepth uint16
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   uint16
	}{
		{
			name: "basic set max entities",
			fields: fields{
				options: &Options{
					MaxEntities: 1,
					MaxDepth:    1,
				},
			},
			args: args{
				maxDepth: 5,
			},
			want: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opFunc := SetMaxDepth(tt.args.maxDepth)
			opFunc(tt.fields.options)
			if tt.fields.options.MaxDepth != tt.want {
				t.Errorf("quadgo.SetMaxEntities() = %v, want %v", tt.fields.options.MaxEntities, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		width, height float64
		ops           []Option
	}
	tests := []struct {
		name string
		args args
		want *QuadGo
	}{
		{
			name: "basic default new",
			args: args{
				width:  800,
				height: 600,
			},
			want: &QuadGo{
				node: &node{
					parent:   nil,
					bound:    NewBound(0, 0, 800, 600),
					entities: make(Entities, 0, defaultOption.MaxEntities),
					children: make(nodes, 0, 4),
					depth:    0,
				},
				maxDepth: defaultOption.MaxDepth,
			},
		},
		{
			name: "new with SetMaxEntities()",
			args: args{
				width:  800,
				height: 600,
				ops: []Option{
					SetMaxEntities(20),
				},
			},
			want: &QuadGo{
				node: &node{
					parent:   nil,
					bound:    NewBound(0, 0, 800, 600),
					entities: make(Entities, 0, 20),
					children: make(nodes, 0, 4),
					depth:    0,
				},
				maxDepth: defaultOption.MaxDepth,
			},
		},
		{
			name: "new with SetMaxDepth()",
			args: args{
				width:  800,
				height: 600,
				ops: []Option{
					SetMaxDepth(10),
				},
			},
			want: &QuadGo{
				node: &node{
					parent:   nil,
					bound:    NewBound(0, 0, 800, 600),
					entities: make(Entities, 0, defaultOption.MaxEntities),
					children: make(nodes, 0, 4),
					depth:    0,
				},
				maxDepth: 10,
			},
		},
		{
			name: "new with SetMaxDepth() and SetMaxEntities()",
			args: args{
				width:  800,
				height: 600,
				ops: []Option{
					SetMaxDepth(10),
					SetMaxEntities(20),
				},
			},
			want: &QuadGo{
				node: &node{
					parent:   nil,
					bound:    NewBound(0, 0, 800, 600),
					entities: make(Entities, 0, 20),
					children: make(nodes, 0, 4),
					depth:    0,
				},
				maxDepth: 10,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.width, tt.args.height, tt.args.ops...)
			if got.maxDepth != tt.want.maxDepth {
				t.Errorf("quadgo.New() for maxDepth = %v, want %v", got.maxDepth, tt.want.maxDepth)
			} else if cap(got.entities) != cap(tt.want.entities) {
				t.Errorf("quadgo.New() for maxEntities = %v, want %v", cap(got.entities), cap(tt.want.entities))
			} else if !got.bound.IsEqual(tt.want.bound) {
				t.Errorf("quadgo.New() for bounds = %v, want %v", got.bound, tt.want.bound)
			}
		})
	}
}

func TestQuadGo_Insert(t *testing.T) {
	type fields struct {
		quadgo *QuadGo
	}
	type args struct {
		minX, minY, maxX, maxY float64
	}
	tests := []struct {
		name   string
		fields fields
		args   []args
		want   Entities
	}{
		{
			name: "basic insert on empty list",
			fields: fields{
				quadgo: New(800, 600),
			},
			args: []args{
				{
					minX: 0,
					minY: 0,
					maxX: 50,
					maxY: 50,
				},
			},
			want: Entities{
				NewEntity(0, 0, 50, 50),
			},
		},
		{
			name: "insert with a split",
			fields: fields{
				quadgo: New(800, 600, SetMaxEntities(2)),
			},
			args: []args{
				{
					minX: 0,
					minY: 0,
					maxX: 50,
					maxY: 50,
				},
				{
					minX: 20,
					minY: 20,
					maxX: 40,
					maxY: 40,
				},
				{
					minX: 25,
					minY: 25,
					maxX: 70,
					maxY: 70,
				},
			},
			want: Entities{
				NewEntity(0, 0, 50, 50),
				NewEntity(20, 20, 40, 40),
				NewEntity(25, 25, 70, 70),
			},
		},
		{
			name: "insert with no split for max depth",
			fields: fields{
				quadgo: New(800, 600, SetMaxEntities(2), SetMaxDepth(0)),
			},
			args: []args{
				{
					minX: 0,
					minY: 0,
					maxX: 50,
					maxY: 50,
				},
				{
					minX: 20,
					minY: 20,
					maxX: 40,
					maxY: 40,
				},
				{
					minX: 25,
					minY: 25,
					maxX: 70,
					maxY: 70,
				},
			},
			want: Entities{
				NewEntity(0, 0, 50, 50),
				NewEntity(20, 20, 40, 40),
				NewEntity(25, 25, 70, 70),
			},
		},
		{
			name: "inert 4 quadrents",
			fields: fields{
				quadgo: New(800, 600, SetMaxEntities(1)),
			},
			args: []args{
				{
					minX: 0,
					minY: 0,
					maxX: 50,
					maxY: 50,
				},
				{
					minX: 0,
					minY: 350,
					maxX: 50,
					maxY: 500,
				},
				{
					minX: 450,
					minY: 0,
					maxX: 600,
					maxY: 50,
				},
				{
					minX: 450,
					minY: 350,
					maxX: 600,
					maxY: 500,
				},
			},
			want: Entities{
				NewEntity(0, 0, 50, 50),
				NewEntity(0, 350, 50, 500),
				NewEntity(450, 0, 600, 50),
				NewEntity(450, 350, 600, 500),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, arg := range tt.args {
				tt.fields.quadgo.Insert(arg.minX, arg.minY, arg.maxX, arg.maxY)
			}

			for _, wnt := range tt.want {
				out := tt.fields.quadgo.IsEntity(wnt)

				got := <-out
				if !got {
					t.Errorf("QuadGo.Insert() could not find %v entity in tree", wnt)
				}
			}
		})
	}
}

func TestQuadGo_InsertWithAction(t *testing.T) {
	type fields struct {
		quadgo *QuadGo
	}
	type args struct {
		minX, minY, maxX, maxY float64
		action                 Action
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Entity
	}{
		{
			name: "insert with action on empty list",
			fields: fields{
				quadgo: New(800, 600),
			},
			args: args{
				minX:   0,
				minY:   0,
				maxX:   50,
				maxY:   50,
				action: func() { fmt.Println("value in a function") },
			},
			want: NewEntityWithAction(0, 0, 50, 50, func() { fmt.Println("value in a function") }),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.quadgo.InsertWithAction(tt.args.minX, tt.args.minY, tt.args.maxX, tt.args.maxY, tt.args.action)

			out := tt.fields.quadgo.IsEntity(tt.want)

			if !<-out {
				t.Errorf("QuadGo.InsertWithAction() entity not found")
			}

			out2 := tt.fields.quadgo.RetrieveFromPoint(tt.want.Center)

			for _, ent := range <-out2 {
				if ent.Action == nil {
					t.Errorf("QuadGo.InsertWithAction() no action set = %v", ent)
				}
			}
		})
	}
}

func TestQuadGo_InsertEntities(t *testing.T) {
	type fields struct {
		quadgo *QuadGo
	}
	type args struct {
		entities Entities
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Entities
		wantErr error
	}{
		{
			name: "insert 1 entity",
			fields: fields{
				quadgo: New(800, 600),
			},
			args: args{
				entities: Entities{
					NewEntity(0, 0, 50, 50),
				},
			},
			want: Entities{
				NewEntity(0, 0, 50, 50),
			},
			wantErr: nil,
		},
		{
			name: "insert no entities error",
			fields: fields{
				quadgo: New(800, 600),
			},
			args: args{
				entities: nil,
			},
			want:    nil,
			wantErr: errors.New("no entities given to QuadGo.InsertEntities()"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fields.quadgo.InsertEntities(tt.args.entities...)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("QuadGo.InsertEntities() unwanted error type = %v, want %v", err, tt.wantErr)
			}

			for _, ent := range tt.want {
				if !<-tt.fields.quadgo.IsEntity(ent) {
					t.Errorf("QuadGo.InsertEntity() could not find entity %v in tree", ent)
				}
			}
		})
	}
}

func TestQuadGo_Remove(t *testing.T) {
	type fields struct {
		quadgo   *QuadGo
		entities Entities
	}
	type args struct {
		entity *Entity
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "remove 1 entity",
			fields: fields{
				quadgo: New(800, 600),
				entities: Entities{
					NewEntity(0, 0, 50, 50),
					NewEntity(20, 20, 50, 50),
					NewEntity(5, 5, 90, 80),
				},
			},
			args: args{
				NewEntity(0, 0, 50, 50),
			},
			wantErr: nil,
		},
		{
			name: "remove and collapse",
			fields: fields{
				quadgo: New(800, 600, SetMaxEntities(2)),
				entities: Entities{
					NewEntity(0, 0, 50, 50),
					NewEntity(20, 20, 50, 50),
					NewEntity(5, 5, 90, 80),
				},
			},
			args: args{
				NewEntity(0, 0, 50, 50),
			},
			wantErr: nil,
		},
		{
			name: "remove non entity error",
			fields: fields{
				quadgo: New(800, 600),
				entities: Entities{
					NewEntity(20, 20, 50, 50),
					NewEntity(5, 5, 90, 80),
				},
			},
			args: args{
				NewEntity(0, 0, 50, 50),
			},
			wantErr: errors.New("could not find entity in tree to remove"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fields.quadgo.InsertEntities(tt.fields.entities...)
			if err != nil {
				t.Errorf("QuadGo.Remove() insert entities with error %v", err)
			}

			err = tt.fields.quadgo.Remove(tt.args.entity)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("QuadGo.Remove() got an unwanted error = %v, want %v", err, tt.wantErr)
			}
			if <-tt.fields.quadgo.IsEntity(tt.args.entity) {
				t.Errorf("QuadGo.Remove() found entity even after delete")
			}
		})
	}
}

func TestQuadGo_RetrieveFromPoint(t *testing.T) {
	type fields struct {
		quadgo   *QuadGo
		entities Entities
	}
	type args struct {
		point Point
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Entities
	}{
		{
			name: "find 1 value",
			fields: fields{
				quadgo: New(800, 600),
				entities: Entities{
					NewEntity(0, 0, 50, 50),
				},
			},
			args: args{
				point: NewPoint(5, 5),
			},
			want: Entities{
				NewEntity(0, 0, 50, 50),
			},
		},
		{
			name: "find 1 value from child",
			fields: fields{
				quadgo: New(800, 600, SetMaxEntities(2)),
				entities: Entities{
					NewEntity(0, 0, 50, 50),
					NewEntity(500, 400, 700, 600),
					NewEntity(450, 350, 600, 550),
				},
			},
			args: args{
				point: NewPoint(5, 5),
			},
			want: Entities{
				NewEntity(0, 0, 50, 50),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fields.quadgo.InsertEntities(tt.fields.entities...)
			if err != nil {
				t.Errorf("QuadGo.RetrieveFromPoint() got error on insert %v", err)
			}

			entities := <-tt.fields.quadgo.RetrieveFromPoint(tt.args.point)

			for _, ent := range tt.want {
				if !entities.Contains(ent) {
					t.Errorf("QuadGo.RetrieveFromPoint() wanted value not found, entities: %v, want: %v", entities, ent)
				}
			}
		})
	}
}

func TestQuadGo_RetrieveFromBound(t *testing.T) {
	type fields struct {
		quadgo   *QuadGo
		entities Entities
	}
	type args struct {
		bound Bound
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Entities
	}{
		{
			name: "find 1 value",
			fields: fields{
				quadgo: New(800, 600),
				entities: Entities{
					NewEntity(0, 0, 50, 50),
				},
			},
			args: args{
				bound: NewBound(5, 5, 10, 10),
			},
			want: Entities{
				NewEntity(0, 0, 50, 50),
			},
		},
		{
			name: "find 1 value from child",
			fields: fields{
				quadgo: New(800, 600, SetMaxEntities(2)),
				entities: Entities{
					NewEntity(0, 0, 50, 50),
					NewEntity(500, 400, 700, 600),
					NewEntity(450, 350, 600, 550),
				},
			},
			args: args{
				bound: NewBound(5, 5, 10, 10),
			},
			want: Entities{
				NewEntity(0, 0, 50, 50),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fields.quadgo.InsertEntities(tt.fields.entities...)
			if err != nil {
				t.Errorf("QuadGo.RetrieveFromBound() got error on insert %v", err)
			}

			entities := <-tt.fields.quadgo.RetrieveFromBound(tt.args.bound)

			for _, ent := range tt.want {
				if !entities.Contains(ent) {
					t.Errorf("QuadGo.RetrieveFromBound() wanted value not found, entities: %v, want: %v", entities, ent)
				}
			}
		})
	}
}

func TestQuadGo_IsEntity(t *testing.T) {
	type fields struct {
		quadgo   *QuadGo
		entities Entities
	}
	type args struct {
		entity *Entity
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "is entity true",
			fields: fields{
				quadgo: New(800, 600),
				entities: Entities{
					NewEntity(0, 0, 50, 50),
				},
			},
			args: args{
				entity: NewEntity(0, 0, 50, 50),
			},
			want: true,
		},
		{
			name: "is entity false",
			fields: fields{
				quadgo: New(800, 600),
				entities: Entities{
					NewEntity(0, 0, 50, 50),
				},
			},
			args: args{
				entity: NewEntity(10, 10, 50, 50),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fields.quadgo.InsertEntities(tt.fields.entities...)
			if err != nil {
				t.Errorf("QuadGo.IsEntity() got error on insert %v", err)
			}

			if got := <-tt.fields.quadgo.IsEntity(tt.args.entity); got != tt.want {
				t.Errorf("QuadGo.IsEntity() = %v, wanted %v", got, tt.want)
			}
		})
	}
}

func TestQuadGo_IsIntersectPoint(t *testing.T) {
	type fields struct {
		quadgo   *QuadGo
		entities Entities
	}
	type args struct {
		point Point
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "is intersect true",
			fields: fields{
				quadgo: New(800, 600),
				entities: Entities{
					NewEntity(0, 0, 50, 50),
				},
			},
			args: args{
				point: NewPoint(5, 5),
			},
			want: true,
		},
		{
			name: "is intersect false",
			fields: fields{
				quadgo: New(800, 600),
				entities: Entities{
					NewEntity(0, 0, 50, 50),
				},
			},
			args: args{
				point: NewPoint(60, 60),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fields.quadgo.InsertEntities(tt.fields.entities...)
			if err != nil {
				t.Errorf("QuadGo.IsIntersectPoint() got error on insert %v", err)
			}

			if got := <-tt.fields.quadgo.IsIntersectPoint(tt.args.point); got != tt.want {
				t.Errorf("QuadGo.IsIntersectPoint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQuadGo_IsIntersectBound(t *testing.T) {
	type fields struct {
		quadgo   *QuadGo
		entities Entities
	}
	type args struct {
		bound Bound
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "is intersect true",
			fields: fields{
				quadgo: New(800, 600),
				entities: Entities{
					NewEntity(0, 0, 50, 50),
				},
			},
			args: args{
				bound: NewBound(5, 5, 10, 10),
			},
			want: true,
		},
		{
			name: "is intersect false",
			fields: fields{
				quadgo: New(800, 600),
				entities: Entities{
					NewEntity(0, 0, 50, 50),
				},
			},
			args: args{
				bound: NewBound(60, 60, 70, 70),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fields.quadgo.InsertEntities(tt.fields.entities...)
			if err != nil {
				t.Errorf("QuadGo.IsIntersectBound() got error on insert %v", err)
			}

			if got := <-tt.fields.quadgo.IsIntersectBound(tt.args.bound); got != tt.want {
				t.Errorf("QuadGo.IsIntersectBound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQuadGo_IntersectsPoint(t *testing.T) {
	type fields struct {
		quadgo   *QuadGo
		entities Entities
	}
	type args struct {
		point Point
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Entities
	}{
		{
			name: "is intersect true",
			fields: fields{
				quadgo: New(800, 600),
				entities: Entities{
					NewEntity(0, 0, 50, 50),
				},
			},
			args: args{
				point: NewPoint(5, 5),
			},
			want: Entities{
				NewEntity(0, 0, 50, 50),
			},
		},
		{
			name: "is intersect false",
			fields: fields{
				quadgo: New(800, 600),
				entities: Entities{
					NewEntity(0, 0, 50, 50),
				},
			},
			args: args{
				point: NewPoint(60, 60),
			},
			want: Entities{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fields.quadgo.InsertEntities(tt.fields.entities...)
			if err != nil {
				t.Errorf("QuadGo.IsIntersectPoint() got error on insert %v", err)
			}

			got := <-tt.fields.quadgo.IntersectsPoint(tt.args.point)

			if len(tt.want) == 0 && len(got) != 0 {
				t.Errorf("QuadGo.IntersectsPoint() wanted no intersects but got %v", got)
			} else {
				for _, ent := range tt.want {
					if !got.Contains(ent) {
						t.Errorf("QuadGo.IntersectsPoint() did not return wanted entity %v", ent)
					}
				}
			}
		})
	}
}

func TestQuadGo_IntersectsBound(t *testing.T) {
	type fields struct {
		quadgo   *QuadGo
		entities Entities
	}
	type args struct {
		bound Bound
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Entities
	}{
		{
			name: "is intersect true",
			fields: fields{
				quadgo: New(800, 600),
				entities: Entities{
					NewEntity(0, 0, 50, 50),
				},
			},
			args: args{
				bound: NewBound(5, 5, 10, 10),
			},
			want: Entities{
				NewEntity(0, 0, 50, 50),
			},
		},
		{
			name: "is intersect false",
			fields: fields{
				quadgo: New(800, 600),
				entities: Entities{
					NewEntity(0, 0, 50, 50),
				},
			},
			args: args{
				bound: NewBound(60, 60, 70, 70),
			},
			want: Entities{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fields.quadgo.InsertEntities(tt.fields.entities...)
			if err != nil {
				t.Errorf("QuadGo.IsIntersectBound() got error on insert %v", err)
			}

			got := <-tt.fields.quadgo.IntersectsBound(tt.args.bound)

			if len(tt.want) == 0 && len(got) != 0 {
				t.Errorf("QuadGo.IntersectsBound() wanted no intersects but got %v", got)
			} else {
				for _, ent := range tt.want {
					if !got.Contains(ent) {
						t.Errorf("QuadGo.IntersectsBound() did not return wanted entity %v", ent)
					}
				}
			}
		})
	}
}
