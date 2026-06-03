package a11y

import (
	"context"
	"fmt"
	"strings"

	"github.com/chromedp/cdproto/accessibility"
	"github.com/chromedp/chromedp"

	"github.com/user/cdp/internal/state"
)

type Element struct {
	UID        string            `json:"uid"`
	Role       string            `json:"role"`
	Name       string            `json:"name,omitempty"`
	Value      string            `json:"value,omitempty"`
	Properties map[string]string `json:"properties,omitempty"`
	Children   []string          `json:"children,omitempty"`
	NodeID     int64             `json:"backendNodeId"`
}

func FetchTree(ctx context.Context, verbose bool) ([]Element, map[string]state.UIDMapping, error) {
	var nodes []*accessibility.Node
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		nodes, err = accessibility.GetFullAXTree().Do(ctx)
		return err
	})); err != nil {
		return nil, nil, err
	}

	var elements []Element
	uidCounter := 0
	uidMap := make(map[string]state.UIDMapping)

	for _, node := range nodes {
		if node.Ignored {
			continue
		}

		role := ""
		if node.Role != nil {
			role = fmt.Sprintf("%v", node.Role.Value)
			role = strings.Trim(role, `"`)
		}

		if role == "" || role == "none" || role == "generic" {
			if !verbose {
				continue
			}
		}

		name := ""
		if node.Name != nil {
			name = fmt.Sprintf("%v", node.Name.Value)
			name = strings.Trim(name, `"`)
		}

		value := ""
		if node.Value != nil {
			value = fmt.Sprintf("%v", node.Value.Value)
			value = strings.Trim(value, `"`)
		}

		uidCounter++
		uid := fmt.Sprintf("a%d", uidCounter)

		props := make(map[string]string)
		for _, p := range node.Properties {
			if p.Value != nil {
				pv := fmt.Sprintf("%v", p.Value.Value)
				pv = strings.Trim(pv, `"`)
				if pv != "" && pv != "false" {
					props[string(p.Name)] = pv
				}
			}
		}

		el := Element{
			UID:        uid,
			Role:       role,
			Name:       name,
			Value:      value,
			Properties: props,
			NodeID:     int64(node.BackendDOMNodeID),
		}
		elements = append(elements, el)

		uidMap[uid] = state.UIDMapping{
			BackendNodeID: int64(node.BackendDOMNodeID),
			Role:          role,
			Name:          name,
		}
	}

	return elements, uidMap, nil
}

func FormatText(elements []Element) string {
	var sb strings.Builder
	for _, el := range elements {
		sb.WriteString(fmt.Sprintf("[%s uid=%s]", el.Role, el.UID))
		if el.Name != "" {
			sb.WriteString(fmt.Sprintf(" \"%s\"", el.Name))
		}
		if el.Value != "" {
			sb.WriteString(fmt.Sprintf(" value=\"%s\"", el.Value))
		}
		for k, v := range el.Properties {
			if k == "focused" || k == "required" || k == "disabled" {
				sb.WriteString(fmt.Sprintf(" %s", k))
			} else {
				sb.WriteString(fmt.Sprintf(" %s=\"%s\"", k, v))
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
