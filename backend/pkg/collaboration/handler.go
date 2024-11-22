package collaboration

import (
	"context"
	"github.com/HardMax71/syncwrite/backend/pkg/document"
	"time"

	"github.com/HardMax71/syncwrite/backend/pkg/auth"
	"github.com/HardMax71/syncwrite/backend/pkg/proto/collaboration/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Handler struct {
	collaborationv1.UnimplementedCollaborationServiceServer
	service         *Service
	documentService *document.Service
}

func NewHandler(service *Service, documentService *document.Service) *Handler {
	return &Handler{
		service:         service,
		documentService: documentService,
	}
}

func (h *Handler) JoinSession(ctx context.Context, req *collaborationv1.JoinSessionRequest) (*collaborationv1.JoinSessionResponse, error) {
	user, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Verify document access
	if _, err := h.documentService.GetDocument(ctx, req.DocumentId, user.ID); err != nil {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	activeUser := &ActiveUser{
		UserID:         user.ID,
		Username:       user.Username,
		CursorPosition: "",
		LastActive:     time.Now(),
	}

	activeUsers, err := h.service.JoinSession(req.DocumentId, activeUser)
	if err != nil {
		return nil, status.Error(codes.Internal, "error joining session")
	}

	protoUsers := make([]*collaborationv1.ActiveUser, len(activeUsers))
	for i, u := range activeUsers {
		protoUsers[i] = &collaborationv1.ActiveUser{
			UserId:         u.UserID,
			Username:       u.Username,
			CursorPosition: u.CursorPosition,
			LastActive:     timestamppb.New(u.LastActive),
		}
	}

	return &collaborationv1.JoinSessionResponse{
		SessionId:   req.DocumentId,
		ActiveUsers: protoUsers,
		MqttTopic:   GetDocumentTopic(req.DocumentId),
	}, nil
}

func (h *Handler) LeaveSession(ctx context.Context, req *collaborationv1.LeaveSessionRequest) (*collaborationv1.LeaveSessionResponse, error) {
	user, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = h.service.LeaveSession(req.DocumentId, user.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "error leaving session")
	}

	return &collaborationv1.LeaveSessionResponse{
		Success: true,
	}, nil
}

func (h *Handler) GetActiveUsers(ctx context.Context, req *collaborationv1.GetActiveUsersRequest) (*collaborationv1.GetActiveUsersResponse, error) {
	users, err := h.service.GetActiveUsers(req.DocumentId)
	if err != nil {
		return nil, status.Error(codes.Internal, "error getting active users")
	}

	protoUsers := make([]*collaborationv1.ActiveUser, len(users))
	for i, u := range users {
		protoUsers[i] = &collaborationv1.ActiveUser{
			UserId:         u.UserID,
			Username:       u.Username,
			CursorPosition: u.CursorPosition,
			LastActive:     timestamppb.New(u.LastActive),
		}
	}

	return &collaborationv1.GetActiveUsersResponse{
		Users: protoUsers,
	}, nil
}

func (h *Handler) StreamChanges(req *collaborationv1.StreamChangesRequest, stream collaborationv1.CollaborationService_StreamChangesServer) error {
	user, err := auth.GetUserFromContext(stream.Context())
	if err != nil {
		return err
	}

	changes, cleanup, err := h.service.StreamChanges(req.DocumentId, user.ID)
	if err != nil {
		return status.Error(codes.Internal, "error setting up change stream")
	}
	defer cleanup()

	for {
		select {
		case change, ok := <-changes:
			if !ok {
				return status.Error(codes.Canceled, "change stream closed")
			}

			protoChange := &collaborationv1.DocumentChange{
				DocumentId: change.DocumentID,
				UserId:     change.UserID,
				Version:    change.Version,
				Timestamp:  timestamppb.New(change.Timestamp),
				Operations: make([]*collaborationv1.Operation, len(change.Operations)),
			}

			for i, op := range change.Operations {
				protoChange.Operations[i] = &collaborationv1.Operation{
					Type:     convertOperationTypeToProto(op.Type),
					Position: op.Position,
					Content:  op.Content,
					Length:   op.Length,
				}
			}

			if err := stream.Send(protoChange); err != nil {
				return status.Error(codes.Internal, "error sending change")
			}

		case <-stream.Context().Done():
			return status.Error(codes.Canceled, "stream context canceled")
		}
	}
}

func (h *Handler) SyncDocument(ctx context.Context, req *collaborationv1.SyncDocumentRequest) (*collaborationv1.SyncDocumentResponse, error) {
	user, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Verify document access
	if _, err := h.documentService.GetDocument(ctx, req.DocumentId, user.ID); err != nil {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	operations := make([]Operation, len(req.Operations))
	for i, op := range req.Operations {
		operations[i] = Operation{
			Type:     convertOperationTypeFromProto(op.Type),
			Position: op.Position,
			Content:  op.Content,
			Length:   op.Length,
		}
	}

	newVersion, concurrentChanges, err := h.service.SyncDocument(ctx, req.DocumentId, operations, req.BaseVersion)
	if err != nil {
		switch err {
		case document.ErrVersionMismatch:
			return nil, status.Error(codes.FailedPrecondition, "version mismatch")
		default:
			return nil, status.Error(codes.Internal, "error syncing document")
		}
	}

	var protoConcurrentChanges []*collaborationv1.DocumentChange
	if concurrentChanges != nil {
		protoConcurrentChanges = make([]*collaborationv1.DocumentChange, len(concurrentChanges))
		for i, change := range concurrentChanges {
			protoOps := make([]*collaborationv1.Operation, len(change.Operations))
			for j, op := range change.Operations {
				protoOps[j] = &collaborationv1.Operation{
					Type:     convertOperationTypeToProto(op.Type),
					Position: op.Position,
					Content:  op.Content,
					Length:   op.Length,
				}
			}

			protoConcurrentChanges[i] = &collaborationv1.DocumentChange{
				DocumentId: change.DocumentID,
				UserId:     change.UserID,
				Version:    change.Version,
				Operations: protoOps,
				Timestamp:  timestamppb.New(change.Timestamp),
			}
		}
	}

	return &collaborationv1.SyncDocumentResponse{
		Success:           true,
		NewVersion:        newVersion,
		ConcurrentChanges: protoConcurrentChanges,
	}, nil
}

// Helper functions for converting between domain and proto types
func convertOperationTypeToProto(t OperationType) collaborationv1.Operation_Type {
	switch t {
	case OperationTypeInsert:
		return collaborationv1.Operation_TYPE_INSERT
	case OperationTypeDelete:
		return collaborationv1.Operation_TYPE_DELETE
	case OperationTypeReplace:
		return collaborationv1.Operation_TYPE_REPLACE
	default:
		return collaborationv1.Operation_TYPE_UNSPECIFIED
	}
}

func convertOperationTypeFromProto(t collaborationv1.Operation_Type) OperationType {
	switch t {
	case collaborationv1.Operation_TYPE_INSERT:
		return OperationTypeInsert
	case collaborationv1.Operation_TYPE_DELETE:
		return OperationTypeDelete
	case collaborationv1.Operation_TYPE_REPLACE:
		return OperationTypeReplace
	default:
		return OperationTypeInsert
	}
}
