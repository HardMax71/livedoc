import { createStore } from 'solid-js/store';
import { Document, DocumentVersion, PermissionLevel } from '../types/document';
import { documentService } from '../api/document';

interface DocumentState {
    documents: Document[];
    currentDocument: Document | null;
    versions: DocumentVersion[];
    total: number;
    isLoading: boolean;
    error: string | null;
}

const initialState: DocumentState = {
    documents: [],
    currentDocument: null,
    versions: [],
    total: 0,
    isLoading: false,
    error: null,
};

const [state, setState] = createStore(initialState);

export const documentStore = {
    get state() {
        return state;
    },

    async listDocuments(page: number = 1, pageSize: number = 10) {
        setState({ isLoading: true, error: null });
        try {
            const { documents, total } = await documentService.listDocuments(page, pageSize);
            setState({ documents, total, isLoading: false });
        } catch (error) {
            setState({
                error: error instanceof Error ? error.message : 'Failed to load documents',
                isLoading: false,
            });
        }
    },

    async getDocument(documentId: string) {
        setState({ isLoading: true, error: null });
        try {
            const document = await documentService.getDocument(documentId);
            setState({ currentDocument: document, isLoading: false });
        } catch (error) {
            setState({
                error: error instanceof Error ? error.message : 'Failed to load document',
                isLoading: false,
            });
        }
    },

    async createDocument(title: string, content: string = '') {
        setState({ isLoading: true, error: null });
        try {
            const document = await documentService.createDocument({ title, content });
            setState(state => ({
                documents: [...state.documents, document],
                currentDocument: document,
                isLoading: false,
            }));
            return document;
        } catch (error) {
            setState({
                error: error instanceof Error ? error.message : 'Failed to create document',
                isLoading: false,
            });
        }
    },

    async updateDocument(documentId: string, title: string, content: string, version: string) {
        setState({ isLoading: true, error: null });
        try {
            const document = await documentService.updateDocument({
                documentId,
                title,
                content,
                version,
            });
            setState(state => ({
                documents: state.documents.map(d =>
                    d.id === document.id ? document : d
                ),
                currentDocument: document,
                isLoading: false,
            }));
            return document;
        } catch (error) {
            setState({
                error: error instanceof Error ? error.message : 'Failed to update document',
                isLoading: false,
            });
        }
    },

    async deleteDocument(documentId: string) {
        setState({ isLoading: true, error: null });
        try {
            await documentService.deleteDocument(documentId);
            setState(state => ({
                documents: state.documents.filter(d => d.id !== documentId),
                currentDocument: null,
                isLoading: false,
            }));
        } catch (error) {
            setState({
                error: error instanceof Error ? error.message : 'Failed to delete document',
                isLoading: false,
            });
        }
    },

    async shareDocument(documentId: string, userEmail: string, permissionLevel: PermissionLevel) {
        setState({ isLoading: true, error: null });
        try {
            await documentService.shareDocument({
                documentId,
                userEmail,
                permissionLevel,
            });
            setState({ isLoading: false });
        } catch (error) {
            setState({
                error: error instanceof Error ? error.message : 'Failed to share document',
                isLoading: false,
            });
        }
    },

    async getVersionHistory(documentId: string, page: number = 1, pageSize: number = 10) {
        setState({ isLoading: true, error: null });
        try {
            const { versions, total } = await documentService.getDocumentHistory(
                documentId,
                page,
                pageSize
            );
            setState({ versions, total, isLoading: false });
        } catch (error) {
            setState({
                error: error instanceof Error ? error.message : 'Failed to load version history',
                isLoading: false,
            });
        }
    },

    async restoreVersion(documentId: string, versionId: string) {
        setState({ isLoading: true, error: null });
        try {
            const document = await documentService.restoreVersion(documentId, versionId);
            setState({
                currentDocument: document,
                isLoading: false,
            });
        } catch (error) {
            setState({
                error: error instanceof Error ? error.message : 'Failed to restore version',
                isLoading: false,
            });
        }
    },

    clearError() {
        setState({ error: null });
    },
};