export interface Document {
    id: string;
    title: string;
    content: string;
    ownerId: string;
    version: string;
    createdAt: Date;
    updatedAt: Date;
}

export interface Permission {
    userId: string;
    documentId: string;
    level: PermissionLevel;
}

export enum PermissionLevel {
    VIEWER = 'VIEWER',
    EDITOR = 'EDITOR',
    OWNER = 'OWNER',
}

export interface DocumentVersion {
    id: string;
    documentId: string;
    content: string;
    editorId: string;
    version: string;
    createdAt: Date;
}

export interface CreateDocumentRequest {
    title: string;
    content: string;
}

export interface UpdateDocumentRequest {
    documentId: string;
    title: string;
    content: string;
    version: string;
}

export interface ShareDocumentRequest {
    documentId: string;
    userEmail: string;
    permissionLevel: PermissionLevel;
}