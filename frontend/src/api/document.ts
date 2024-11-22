import { createClient, getAuthHeader } from './grpc-client';
import type {
    Document,
    CreateDocumentRequest,
    UpdateDocumentRequest,
    ShareDocumentRequest,
    DocumentVersion
} from '../types/document';

class DocumentService {
    private client = createClient(DocumentService);

    async createDocument(request: CreateDocumentRequest): Promise<Document> {
        const response = await this.client.createDocument(request, {
            headers: getAuthHeader(),
        });
        return response.document;
    }

    async getDocument(documentId: string): Promise<Document> {
        const response = await this.client.getDocument(
            { documentId },
            { headers: getAuthHeader() }
        );
        return response.document;
    }

    async updateDocument(request: UpdateDocumentRequest): Promise<Document> {
        const response = await this.client.updateDocument(request, {
            headers: getAuthHeader(),
        });
        return response.document;
    }

    async deleteDocument(documentId: string): Promise<boolean> {
        const response = await this.client.deleteDocument(
            { documentId },
            { headers: getAuthHeader() }
        );
        return response.success;
    }

    async listDocuments(page: number = 1, pageSize: number = 10): Promise<{
        documents: Document[];
        total: number;
    }> {
        const response = await this.client.listDocuments(
            { page, pageSize },
            { headers: getAuthHeader() }
        );
        return {
            documents: response.documents,
            total: response.total,
        };
    }

    async shareDocument(request: ShareDocumentRequest): Promise<boolean> {
        const response = await this.client.shareDocument(request, {
            headers: getAuthHeader(),
        });
        return response.success;
    }

    async getDocumentHistory(
        documentId: string,
        page: number = 1,
        pageSize: number = 10
    ): Promise<{
        versions: DocumentVersion[];
        total: number;
    }> {
        const response = await this.client.getDocumentHistory(
            { documentId, page, pageSize },
            { headers: getAuthHeader() }
        );
        return {
            versions: response.versions,
            total: response.total,
        };
    }

    async restoreVersion(documentId: string, versionId: string): Promise<Document> {
        const response = await this.client.restoreVersion(
            { documentId, versionId },
            { headers: getAuthHeader() }
        );
        return response.document;
    }
}

export const documentService = new DocumentService();