package document

import (
	"context"
	"github.com/HardMax71/syncwrite/backend/pkg/auth"
	"github.com/HardMax71/syncwrite/backend/pkg/proto/document/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Handler struct {
	documentv1.UnimplementedDocumentServiceServer
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateDocument(ctx context.Context, req *documentv1.CreateDocumentRequest) (*documentv1.DocumentResponse, error) {
	user, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	params := CreateDocumentParams{
		Title:   req.Title,
		Content: req.Content,
		OwnerID: user.ID,
	}

	doc, err := h.service.CreateDocument(ctx, params)
	if err != nil {
		return nil, status.Error(codes.Internal, "error creating document")
	}

	return &documentv1.DocumentResponse{
		Document: convertDocumentToProto(doc),
	}, nil
}

func (h *Handler) GetDocument(ctx context.Context, req *documentv1.GetDocumentRequest) (*documentv1.DocumentResponse, error) {
	user, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	doc, err := h.service.GetDocument(ctx, req.DocumentId, user.ID)
	if err != nil {
		switch err {
		case ErrDocumentNotFound:
			return nil, status.Error(codes.NotFound, "document not found")
		case ErrPermissionDenied:
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &documentv1.DocumentResponse{
		Document: convertDocumentToProto(doc),
	}, nil
}

func (h *Handler) UpdateDocument(ctx context.Context, req *documentv1.UpdateDocumentRequest) (*documentv1.DocumentResponse, error) {
	user, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	params := UpdateDocumentParams{
		DocumentID: req.DocumentId,
		Title:      req.Title,
		Content:    req.Content,
		Version:    req.Version,
		EditorID:   user.ID,
	}

	doc, err := h.service.UpdateDocument(ctx, params)
	if err != nil {
		switch err {
		case ErrDocumentNotFound:
			return nil, status.Error(codes.NotFound, "document not found")
		case ErrVersionMismatch:
			return nil, status.Error(codes.FailedPrecondition, "version mismatch")
		case ErrPermissionDenied:
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &documentv1.DocumentResponse{
		Document: convertDocumentToProto(doc),
	}, nil
}

func (h *Handler) DeleteDocument(ctx context.Context, req *documentv1.DeleteDocumentRequest) (*documentv1.DeleteDocumentResponse, error) {
	user, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = h.service.DeleteDocument(ctx, req.DocumentId, user.ID)
	if err != nil {
		switch err {
		case ErrDocumentNotFound:
			return nil, status.Error(codes.NotFound, "document not found")
		case ErrPermissionDenied:
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &documentv1.DeleteDocumentResponse{
		Success: true,
	}, nil
}

func (h *Handler) ListDocuments(ctx context.Context, req *documentv1.ListDocumentsRequest) (*documentv1.ListDocumentsResponse, error) {
	user, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	documents, total, err := h.service.ListDocuments(ctx, user.ID, req.Page, req.PageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, "error listing documents")
	}

	protoDocuments := make([]*documentv1.Document, len(documents))
	for i, doc := range documents {
		protoDocuments[i] = convertDocumentToProto(doc)
	}

	return &documentv1.ListDocumentsResponse{
		Documents: protoDocuments,
		Total:     total,
	}, nil
}

func (h *Handler) ShareDocument(ctx context.Context, req *documentv1.ShareDocumentRequest) (*documentv1.ShareDocumentResponse, error) {
	user, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Verify that the user has owner permissions
	doc, err := h.service.GetDocument(ctx, req.DocumentId, user.ID)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	if doc.OwnerID != user.ID {
		return nil, status.Error(codes.PermissionDenied, "only document owner can share the document")
	}

	params := ShareDocumentParams{
		DocumentID: req.DocumentId,
		UserID:     req.UserEmail,
		Level:      convertPermissionLevelFromProto(req.PermissionLevel),
	}

	permission, err := h.service.ShareDocument(ctx, params)
	if err != nil {
		switch err {
		case ErrInvalidPermission:
			return nil, status.Error(codes.InvalidArgument, "invalid permission level")
		case ErrDocumentNotFound:
			return nil, status.Error(codes.NotFound, "document not found")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &documentv1.ShareDocumentResponse{
		Success: true,
		Permission: &documentv1.Permission{
			UserId:     permission.UserID,
			DocumentId: permission.DocumentID,
			Level:      convertPermissionLevelToProto(permission.Level),
		},
	}, nil
}

func (h *Handler) GetDocumentHistory(ctx context.Context, req *documentv1.GetDocumentHistoryRequest) (*documentv1.GetDocumentHistoryResponse, error) {
	user, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Check document access first
	if _, err := h.service.GetDocument(ctx, req.DocumentId, user.ID); err != nil {
		switch err {
		case ErrDocumentNotFound:
			return nil, status.Error(codes.NotFound, "document not found")
		case ErrPermissionDenied:
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	versions, total, err := h.service.GetDocumentHistory(ctx, req.DocumentId, req.Page, req.PageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, "error retrieving document history")
	}

	protoVersions := make([]*documentv1.DocumentVersion, len(versions))
	for i, version := range versions {
		protoVersions[i] = &documentv1.DocumentVersion{
			Id:         version.ID,
			DocumentId: version.DocumentID,
			Content:    version.Content,
			EditorId:   version.EditorID,
			CreatedAt:  timestamppb.New(version.CreatedAt),
		}
	}

	return &documentv1.GetDocumentHistoryResponse{
		Versions: protoVersions,
		Total:    total,
	}, nil
}

func (h *Handler) RestoreVersion(ctx context.Context, req *documentv1.RestoreVersionRequest) (*documentv1.DocumentResponse, error) {
	user, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	doc, err := h.service.RestoreVersion(ctx, req.DocumentId, req.VersionId, user.ID)
	if err != nil {
		switch err {
		case ErrDocumentNotFound:
			return nil, status.Error(codes.NotFound, "document not found")
		case ErrPermissionDenied:
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &documentv1.DocumentResponse{
		Document: convertDocumentToProto(doc),
	}, nil
}

// Helper functions for converting between domain and proto types
func convertDocumentToProto(doc *Document) *documentv1.Document {
	return &documentv1.Document{
		Id:        doc.ID,
		Title:     doc.Title,
		Content:   doc.Content,
		OwnerId:   doc.OwnerID,
		Version:   doc.Version,
		CreatedAt: timestamppb.New(doc.CreatedAt),
		UpdatedAt: timestamppb.New(doc.UpdatedAt),
	}
}

func convertPermissionLevelToProto(level string) documentv1.PermissionLevel {
	switch level {
	case PermissionLevelViewer:
		return documentv1.PermissionLevel_PERMISSION_LEVEL_VIEWER
	case PermissionLevelEditor:
		return documentv1.PermissionLevel_PERMISSION_LEVEL_EDITOR
	case PermissionLevelOwner:
		return documentv1.PermissionLevel_PERMISSION_LEVEL_OWNER
	default:
		return documentv1.PermissionLevel_PERMISSION_LEVEL_UNSPECIFIED
	}
}

func convertPermissionLevelFromProto(level documentv1.PermissionLevel) string {
	switch level {
	case documentv1.PermissionLevel_PERMISSION_LEVEL_VIEWER:
		return PermissionLevelViewer
	case documentv1.PermissionLevel_PERMISSION_LEVEL_EDITOR:
		return PermissionLevelEditor
	case documentv1.PermissionLevel_PERMISSION_LEVEL_OWNER:
		return PermissionLevelOwner
	default:
		return ""
	}
}
