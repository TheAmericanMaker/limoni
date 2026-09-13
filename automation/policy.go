package automation

import "github.com/thebanri/limoni/core/accessibility"

// Policy decides what the server lets leave the process.
//
// Every field defaults to the closed position. The server is a gateway between
// a running application and whatever connects to its socket, and the safe
// default for a gateway is to expose structure — roles, labels, positions,
// bounds — and nothing a user typed or anything the application did not mark as
// safe to show.
type Policy struct {
	// ExposeInputValues sends the value of input nodes (text fields, text areas).
	// Off by default because a text field is where a user types things that
	// matter: an email, a search, a card number that the application did not
	// think to mark Secret.
	//
	// Nodes marked StateSensitive are never exposed, whatever this says.
	ExposeInputValues bool

	// ExposeScreen allows the snapshot op, which returns the rendered grid.
	// Off by default because the grid contains every character on screen, and
	// the application cannot mark an arbitrary paragraph as secret.
	ExposeScreen bool

	// AllowInput permits key, text and click synthesis. Off by default: with it
	// off, a client can observe the application but cannot act on it.
	AllowInput bool

	// AllowUnverifiedPeers accepts connections whose owning user the kernel
	// cannot report. Linux, macOS and FreeBSD report it, and there a connection
	// from any other user is always refused. Elsewhere — Windows among them —
	// the socket's file permissions are the only protection, so the server
	// refuses every connection unless this is set.
	AllowUnverifiedPeers bool
}

// redactTree returns a copy of the tree with everything the policy withholds
// removed. The input is not modified: it is the frame's tree, still owned by
// the application.
func (p Policy) redactTree(nodes []accessibility.AccessibilityNode) []accessibility.AccessibilityNode {
	if nodes == nil {
		return nil
	}
	out := make([]accessibility.AccessibilityNode, len(nodes))
	for i, node := range nodes {
		out[i] = p.redactNode(node)
	}
	return out
}

func (p Policy) redactNode(node accessibility.AccessibilityNode) accessibility.AccessibilityNode {
	// Unconditional. A widget that sets StateSensitive is meant to leave Value
	// empty itself; clearing it here as well means one forgotten line in a
	// widget does not become a leaked password.
	if node.State&accessibility.StateSensitive != 0 {
		node.Value = ""
		node.Description = ""
	} else if node.Role == accessibility.RoleInput && !p.ExposeInputValues {
		node.Value = ""
	}
	node.Children = p.redactTree(node.Children)
	return node
}
