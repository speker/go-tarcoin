package reads_test

import (
	"testing"

	"github.com/influxdata/influxdb/storage/reads"
	"github.com/influxdata/influxdb/storage/reads/datatypes"
)

func TestHasFieldValueKey(t *testing.T) {
	predicates := []*datatypes.Node{
		{
			NodeType: datatypes.NodeTypeComparisonExpression,
			Value: &datatypes.Node_Comparison_{
				Comparison: datatypes.ComparisonLess,
			},
			Children: []*datatypes.Node{
				{
					NodeType: datatypes.NodeTypeFieldRef,
					Value: &datatypes.Node_FieldRefValue{
						FieldRefValue: "_value",
					},
				},
				{
					NodeType: datatypes.NodeTypeLiteral,
					Value: &datatypes.Node_IntegerValue{
						IntegerValue: 3000,
					},
				},
			},
		},
		{
			NodeType: datatypes.NodeTypeLogicalExpression,
			Value: &datatypes.Node_Logical_{
				Logical: datatypes.LogicalAnd,
			},
			Children: []*datatypes.Node{
				{
					NodeType: datatypes.NodeTypeComparisonExpression,
					Value: &datatypes.Node_Comparison_{
						Comparison: datatypes.ComparisonEqual,
					},
					Children: []*datatypes.Node{
						{
							NodeType: datatypes.NodeTypeTagRef,
							Value: &datatypes.Node_TagRefValue{
								TagRefValue: "_measurement",
							},
						},
						{
							NodeType: datatypes.NodeTypeLiteral,
							Value: &datatypes.Node_StringValue{
								StringValue: "cpu",
							},
						},
					},
				},
				{
					NodeType: datatypes.NodeTypeComparisonExpression,
					Value: &datatypes.Node_Comparison_{
						Comparison: datatypes.ComparisonLess,
					},
					Children: []*datatypes.Node{
						{
							NodeType: datatypes.NodeTypeFieldRef,
							Value: &datatypes.Node_FieldRefValue{
								FieldRefValue: "_value",
							},
						},
						{
							NodeType: datatypes.NodeTypeLiteral,
							Value: &datatypes.Node_IntegerValue{
								IntegerValue: 3000,
							},
						},
					},
				},
			},
		},
	}
	for _, predicate := range predicates {
		t.Run("", func(t *testing.T) {
			expr, err := reads.NodeToExpr(predicate, nil)
			if err != nil {
				t.Fatalf("unexpected error converting predicate to InfluxQL expression: %v", err)
			}
			if !reads.HasFieldValueKey(expr) {
				t.Fatalf("did not find a field reference in %v", expr)
			}
		})
	}
}
