export interface ActiveUser {
    userId: string;
    username: string;
    cursorPosition: string;
    lastActive: Date;
}

export interface Operation {
    type: OperationType;
    position: number;
    content: string;
    length: number;
}

export enum OperationType {
    INSERT = 'INSERT',
    DELETE = 'DELETE',
    REPLACE = 'REPLACE',
}

export interface DocumentChange {
    documentId: string;
    userId: string;
    version: string;
    operations: Operation[];
    timestamp: Date;
}

export interface JoinSessionResponse {
    sessionId: string;
    activeUsers: ActiveUser[];
    mqttTopic: string;
}

export interface SyncDocumentResponse {
    success: boolean;
    newVersion: string;
    concurrentChanges: DocumentChange[];
}