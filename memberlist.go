package grkv

import (
	"github.com/hashicorp/memberlist"
	"go.uber.org/zap"
)

// NotifyJoin logs new node joining event.
func (s *Store) NotifyJoin(node *memberlist.Node) {
	s.logger.Info("node joined", zap.Any("node_joined", node.Name), zap.Any("metadata", node.Meta))
	s.addNode(node.Name, node.Addr.String())

}

// NotifyLeave logs new node leave event.
func (s *Store) NotifyLeave(node *memberlist.Node) {
	s.logger.Info("node left", zap.Any("node", node))
	go s.removeNode(node.Name)

}

// NotifyUpdate logs node update events
func (s *Store) NotifyUpdate(node *memberlist.Node) {
	s.logger.Info("node updated", zap.Any("node", node))
}
