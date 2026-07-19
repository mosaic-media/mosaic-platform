package graphql

import (
	"github.com/graphql-go/graphql"

	v1 "github.com/mosaic-media/mosaic-sdk/contracts/platform/v1"

	"github.com/mosaic-media/mosaic-platform/internal/platform/app"
)

// The content projection surface. Every resolver here calls exactly one
// app.Service content method, passing a v1.Caller built from the request's
// session argument (ADR 0017) — the same opaque reference a compiled-in
// capability forwards. Content models come from the published SDK
// (contracts/platform/v1); this package maps them to GraphQL and nothing more.

// caller builds the v1.Caller a content command or query carries from the
// required callerSessionId argument.
func caller(p graphql.ResolveParams) v1.Caller {
	return v1.CallerFromSession(argString(p, "callerSessionId"))
}

func argString(p graphql.ResolveParams, name string) string {
	if v, ok := p.Args[name].(string); ok {
		return v
	}
	return ""
}

func argFloat(p graphql.ResolveParams, name string) float64 {
	if v, ok := p.Args[name].(float64); ok {
		return v
	}
	return 0
}

func argInt(p graphql.ResolveParams, name string) int {
	if v, ok := p.Args[name].(int); ok {
		return v
	}
	return 0
}

func argBool(p graphql.ResolveParams, name string) bool {
	v, _ := p.Args[name].(bool)
	return v
}

// optionalBytes maps an absent-or-empty JSON string argument to nil, so the
// service applies its empty-document default rather than storing "".
func optionalBytes(s string) []byte {
	if s == "" {
		return nil
	}
	return []byte(s)
}

// nodeType projects v1.Node. Every field resolves explicitly because the SDK
// uses named string types and raw JSON []byte, which graphql-go's reflection
// default does not serialize.
var nodeType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Node",
	Fields: graphql.Fields{
		"id":            strField(func(n v1.Node) string { return string(n.ID) }),
		"workId":        strField(func(n v1.Node) string { return string(n.WorkID) }),
		"parentId":      strField(func(n v1.Node) string { return nodeIDPtr(n.ParentID) }),
		"kind":          strField(func(n v1.Node) string { return string(n.Kind) }),
		"mediaType":     strField(func(n v1.Node) string { return string(n.MediaType) }),
		"containerType": strField(func(n v1.Node) string { return string(n.ContainerType) }),
		"itemType":      strField(func(n v1.Node) string { return string(n.ItemType) }),
		"title":         strField(func(n v1.Node) string { return n.Title }),
		"status":        strField(func(n v1.Node) string { return string(n.Status) }),
		"externalIds":   strField(func(n v1.Node) string { return string(n.ExternalIDs) }),
		"attributes":    strField(func(n v1.Node) string { return string(n.Attributes) }),
		"naturalOrder": &graphql.Field{Type: graphql.Float, Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return p.Source.(v1.Node).NaturalOrder, nil
		}},
		"createdAt": &graphql.Field{Type: graphql.String, Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return formatTime(p.Source.(v1.Node).CreatedAt), nil
		}},
		"updatedAt": &graphql.Field{Type: graphql.String, Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return formatTime(p.Source.(v1.Node).UpdatedAt), nil
		}},
	},
})

// partType projects v1.Part.
var partType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Part",
	Fields: graphql.Fields{
		"id":               partStr(func(p v1.Part) string { return string(p.ID) }),
		"nodeId":           partStr(func(p v1.Part) string { return string(p.NodeID) }),
		"role":             partStr(func(p v1.Part) string { return string(p.Role) }),
		"editionLabel":     partStr(func(p v1.Part) string { return p.EditionLabel }),
		"locationScheme":   partStr(func(p v1.Part) string { return string(p.Location.Scheme) }),
		"locationProvider": partStr(func(p v1.Part) string { return p.Location.Provider }),
		"locationRef":      partStr(func(p v1.Part) string { return p.Location.Ref }),
		"container":        partStr(func(p v1.Part) string { return p.Container }),
		"videoCodec":       partStr(func(p v1.Part) string { return p.VideoCodec }),
		"audioCodec":       partStr(func(p v1.Part) string { return p.AudioCodec }),
		"hdrFormat":        partStr(func(p v1.Part) string { return p.HDRFormat }),
		"naturalOrder": &graphql.Field{Type: graphql.Float, Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return p.Source.(v1.Part).NaturalOrder, nil
		}},
		"width":  &graphql.Field{Type: graphql.Int, Resolve: func(p graphql.ResolveParams) (interface{}, error) { return p.Source.(v1.Part).Width, nil }},
		"height": &graphql.Field{Type: graphql.Int, Resolve: func(p graphql.ResolveParams) (interface{}, error) { return p.Source.(v1.Part).Height, nil }},
		"durationSeconds": &graphql.Field{Type: graphql.Float, Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return p.Source.(v1.Part).Duration.Seconds(), nil
		}},
		"bitrateBps": &graphql.Field{Type: graphql.Float, Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return float64(p.Source.(v1.Part).BitrateBPS), nil
		}},
		"sizeBytes": &graphql.Field{Type: graphql.Float, Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return float64(p.Source.(v1.Part).SizeBytes), nil
		}},
	},
})

// relationType projects v1.Relation.
var relationType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Relation",
	Fields: graphql.Fields{
		"id":         relStr(func(r v1.Relation) string { return string(r.ID) }),
		"fromNodeId": relStr(func(r v1.Relation) string { return string(r.FromNodeID) }),
		"toNodeId":   relStr(func(r v1.Relation) string { return string(r.ToNodeID) }),
		"type":       relStr(func(r v1.Relation) string { return string(r.Type) }),
		"origin":     relStr(func(r v1.Relation) string { return string(r.Origin) }),
		"confidence": &graphql.Field{Type: graphql.Float, Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return p.Source.(v1.Relation).Confidence, nil
		}},
		"createdAt": &graphql.Field{Type: graphql.String, Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return formatTime(p.Source.(v1.Relation).CreatedAt), nil
		}},
	},
})

// sourceBindingType projects v1.SourceBinding.
var sourceBindingType = graphql.NewObject(graphql.ObjectConfig{
	Name: "SourceBinding",
	Fields: graphql.Fields{
		"id":             bindStr(func(b v1.SourceBinding) string { return string(b.ID) }),
		"nodeId":         bindStr(func(b v1.SourceBinding) string { return string(b.NodeID) }),
		"sourceProvider": bindStr(func(b v1.SourceBinding) string { return b.SourceProvider }),
		"sourceRef":      bindStr(func(b v1.SourceBinding) string { return b.SourceRef }),
		"matchMethod":    bindStr(func(b v1.SourceBinding) string { return string(b.MatchMethod) }),
		"status":         bindStr(func(b v1.SourceBinding) string { return string(b.Status) }),
		"matchConfidence": &graphql.Field{Type: graphql.Float, Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return p.Source.(v1.SourceBinding).MatchConfidence, nil
		}},
	},
})

// contentNodePayloadType wraps GetContentNodeResult: a node plus, when asked,
// its direct children.
var contentNodePayloadType = graphql.NewObject(graphql.ObjectConfig{
	Name: "ContentNodePayload",
	Fields: graphql.Fields{
		"node":     &graphql.Field{Type: nodeType},
		"children": &graphql.Field{Type: graphql.NewList(nodeType)},
	},
})

// ---- field-resolver helpers, one per source type ----

func strField(get func(v1.Node) string) *graphql.Field {
	return &graphql.Field{Type: graphql.String, Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		return get(p.Source.(v1.Node)), nil
	}}
}
func partStr(get func(v1.Part) string) *graphql.Field {
	return &graphql.Field{Type: graphql.String, Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		return get(p.Source.(v1.Part)), nil
	}}
}
func relStr(get func(v1.Relation) string) *graphql.Field {
	return &graphql.Field{Type: graphql.String, Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		return get(p.Source.(v1.Relation)), nil
	}}
}
func bindStr(get func(v1.SourceBinding) string) *graphql.Field {
	return &graphql.Field{Type: graphql.String, Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		return get(p.Source.(v1.SourceBinding)), nil
	}}
}

func nodeIDPtr(id *v1.NodeID) string {
	if id == nil {
		return ""
	}
	return string(*id)
}

// ---- queries ----

func searchContentField(svc *app.Service) *graphql.Field {
	return &graphql.Field{
		Type: graphql.NewList(nodeType),
		Args: graphql.FieldConfigArgument{
			"callerSessionId": nonNullString(),
			"title":           &graphql.ArgumentConfig{Type: graphql.String},
			"mediaType":       &graphql.ArgumentConfig{Type: graphql.String},
			"kind":            &graphql.ArgumentConfig{Type: graphql.String},
			"limit":           &graphql.ArgumentConfig{Type: graphql.Int},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			result, err := svc.SearchContent(p.Context, v1.SearchContentQuery{
				Caller:    caller(p),
				Title:     argString(p, "title"),
				MediaType: v1.MediaType(argString(p, "mediaType")),
				Kind:      v1.NodeKind(argString(p, "kind")),
				Limit:     argInt(p, "limit"),
			})
			if err != nil {
				return nil, err
			}
			return result.Nodes, nil
		},
	}
}

func contentNodeField(svc *app.Service) *graphql.Field {
	return &graphql.Field{
		Type: contentNodePayloadType,
		Args: graphql.FieldConfigArgument{
			"callerSessionId": nonNullString(),
			"id":              nonNullString(),
			"withChildren":    &graphql.ArgumentConfig{Type: graphql.Boolean},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			result, err := svc.GetContentNode(p.Context, v1.GetContentNodeQuery{
				Caller:       caller(p),
				NodeID:       v1.NodeID(argString(p, "id")),
				WithChildren: argBool(p, "withChildren"),
			})
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{"node": result.Node, "children": result.Children}, nil
		},
	}
}

func contentByExternalIDField(svc *app.Service) *graphql.Field {
	return &graphql.Field{
		Type: graphql.NewList(nodeType),
		Args: graphql.FieldConfigArgument{
			"callerSessionId": nonNullString(),
			"scheme":          nonNullString(),
			"value":           nonNullString(),
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			result, err := svc.FindContentByExternalID(p.Context, v1.FindContentByExternalIDQuery{
				Caller: caller(p),
				Scheme: argString(p, "scheme"),
				Value:  argString(p, "value"),
			})
			if err != nil {
				return nil, err
			}
			return result.Nodes, nil
		},
	}
}

// ---- mutations ----

func addContentWorkField(svc *app.Service) *graphql.Field {
	return &graphql.Field{
		Type: nodeType,
		Args: graphql.FieldConfigArgument{
			"callerSessionId": nonNullString(),
			"mediaType":       nonNullString(),
			"title":           nonNullString(),
			"externalIds":     &graphql.ArgumentConfig{Type: graphql.String},
			"attributes":      &graphql.ArgumentConfig{Type: graphql.String},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			result, err := svc.AddContentWork(p.Context, v1.AddContentWorkCommand{
				Caller:      caller(p),
				MediaType:   v1.MediaType(argString(p, "mediaType")),
				Title:       argString(p, "title"),
				ExternalIDs: optionalBytes(argString(p, "externalIds")),
				Attributes:  optionalBytes(argString(p, "attributes")),
			})
			if err != nil {
				return nil, err
			}
			return result.Work, nil
		},
	}
}

func addContentChildField(svc *app.Service) *graphql.Field {
	return &graphql.Field{
		Type: nodeType,
		Args: graphql.FieldConfigArgument{
			"callerSessionId": nonNullString(),
			"parentId":        nonNullString(),
			"kind":            nonNullString(),
			"containerType":   &graphql.ArgumentConfig{Type: graphql.String},
			"itemType":        &graphql.ArgumentConfig{Type: graphql.String},
			"title":           nonNullString(),
			"naturalOrder":    &graphql.ArgumentConfig{Type: graphql.Float},
			"externalIds":     &graphql.ArgumentConfig{Type: graphql.String},
			"attributes":      &graphql.ArgumentConfig{Type: graphql.String},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			result, err := svc.AddContentChild(p.Context, v1.AddContentChildCommand{
				Caller:        caller(p),
				ParentID:      v1.NodeID(argString(p, "parentId")),
				Kind:          v1.NodeKind(argString(p, "kind")),
				ContainerType: v1.ContainerType(argString(p, "containerType")),
				ItemType:      v1.ItemType(argString(p, "itemType")),
				Title:         argString(p, "title"),
				NaturalOrder:  argFloat(p, "naturalOrder"),
				ExternalIDs:   optionalBytes(argString(p, "externalIds")),
				Attributes:    optionalBytes(argString(p, "attributes")),
			})
			if err != nil {
				return nil, err
			}
			return result.Node, nil
		},
	}
}

func attachContentPartField(svc *app.Service) *graphql.Field {
	return &graphql.Field{
		Type: partType,
		Args: graphql.FieldConfigArgument{
			"callerSessionId":  nonNullString(),
			"nodeId":           nonNullString(),
			"role":             nonNullString(),
			"editionLabel":     &graphql.ArgumentConfig{Type: graphql.String},
			"naturalOrder":     &graphql.ArgumentConfig{Type: graphql.Float},
			"locationScheme":   nonNullString(),
			"locationProvider": &graphql.ArgumentConfig{Type: graphql.String},
			"locationRef":      nonNullString(),
			"container":        &graphql.ArgumentConfig{Type: graphql.String},
			"videoCodec":       &graphql.ArgumentConfig{Type: graphql.String},
			"audioCodec":       &graphql.ArgumentConfig{Type: graphql.String},
			"width":            &graphql.ArgumentConfig{Type: graphql.Int},
			"height":           &graphql.ArgumentConfig{Type: graphql.Int},
			"hdrFormat":        &graphql.ArgumentConfig{Type: graphql.String},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			result, err := svc.AttachContentPart(p.Context, v1.AttachContentPartCommand{
				Caller:       caller(p),
				NodeID:       v1.NodeID(argString(p, "nodeId")),
				Role:         v1.PartRole(argString(p, "role")),
				EditionLabel: argString(p, "editionLabel"),
				NaturalOrder: argFloat(p, "naturalOrder"),
				Location: v1.MediaLocation{
					Scheme:   v1.LocationScheme(argString(p, "locationScheme")),
					Provider: argString(p, "locationProvider"),
					Ref:      argString(p, "locationRef"),
				},
				Container:  argString(p, "container"),
				VideoCodec: argString(p, "videoCodec"),
				AudioCodec: argString(p, "audioCodec"),
				Width:      argInt(p, "width"),
				Height:     argInt(p, "height"),
				HDRFormat:  argString(p, "hdrFormat"),
			})
			if err != nil {
				return nil, err
			}
			return result.Part, nil
		},
	}
}

func relateContentField(svc *app.Service) *graphql.Field {
	return &graphql.Field{
		Type: relationType,
		Args: graphql.FieldConfigArgument{
			"callerSessionId": nonNullString(),
			"fromNodeId":      nonNullString(),
			"toNodeId":        nonNullString(),
			"type":            nonNullString(),
			"confidence":      &graphql.ArgumentConfig{Type: graphql.Float},
			"origin":          nonNullString(),
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			result, err := svc.RelateContent(p.Context, v1.RelateContentCommand{
				Caller:     caller(p),
				FromNodeID: v1.NodeID(argString(p, "fromNodeId")),
				ToNodeID:   v1.NodeID(argString(p, "toNodeId")),
				Type:       v1.RelationType(argString(p, "type")),
				Confidence: argFloat(p, "confidence"),
				Origin:     v1.RelationOrigin(argString(p, "origin")),
			})
			if err != nil {
				return nil, err
			}
			return result.Relation, nil
		},
	}
}

func bindContentSourceField(svc *app.Service) *graphql.Field {
	return &graphql.Field{
		Type: sourceBindingType,
		Args: graphql.FieldConfigArgument{
			"callerSessionId": nonNullString(),
			"nodeId":          nonNullString(),
			"sourceProvider":  nonNullString(),
			"sourceRef":       nonNullString(),
			"matchConfidence": &graphql.ArgumentConfig{Type: graphql.Float},
			"matchMethod":     nonNullString(),
			"status":          nonNullString(),
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			result, err := svc.BindContentSource(p.Context, v1.BindContentSourceCommand{
				Caller:          caller(p),
				NodeID:          v1.NodeID(argString(p, "nodeId")),
				SourceProvider:  argString(p, "sourceProvider"),
				SourceRef:       argString(p, "sourceRef"),
				MatchConfidence: argFloat(p, "matchConfidence"),
				MatchMethod:     v1.MatchMethod(argString(p, "matchMethod")),
				Status:          v1.BindingStatus(argString(p, "status")),
			})
			if err != nil {
				return nil, err
			}
			return result.Binding, nil
		},
	}
}

func resolveContentBindingField(svc *app.Service) *graphql.Field {
	return &graphql.Field{
		Type: sourceBindingType,
		Args: graphql.FieldConfigArgument{
			"callerSessionId": nonNullString(),
			"bindingId":       nonNullString(),
			"resolution":      nonNullString(),
			"moveToNodeId":    &graphql.ArgumentConfig{Type: graphql.String},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			result, err := svc.ResolveContentBinding(p.Context, v1.ResolveContentBindingCommand{
				Caller:       caller(p),
				BindingID:    v1.SourceBindingID(argString(p, "bindingId")),
				Resolution:   v1.BindingResolution(argString(p, "resolution")),
				MoveToNodeID: v1.NodeID(argString(p, "moveToNodeId")),
			})
			if err != nil {
				return nil, err
			}
			return result.Binding, nil
		},
	}
}

func nonNullString() *graphql.ArgumentConfig {
	return &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)}
}
