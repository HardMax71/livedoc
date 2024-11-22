import { createClient, getAuthHeader } from './grpc-client';
import type {
    ActiveUser,
    DocumentChange,
    JoinSessionResponse,
    Operation,
    SyncDocumentResponse
} from '../types/collaboration';
import { connect } from 'mqtt';

class CollaborationService {
    private client = createClient(CollaborationService);
    private mqttClient: any;
    private subscriptions: Map<string, (change: DocumentChange) => void> = new Map();

    async joinSession(documentId: string): Promise<JoinSessionResponse> {
        const response = await this.client.joinSession(
            { documentId },
            { headers: getAuthHeader() }
        );
        await this.connectMqtt(response.mqttTopic);
        return response;
    }

    async leaveSession(sessionId: string, documentId: string): Promise<boolean> {
        const response = await this.client.leaveSession(
            { sessionId, documentId },
            { headers: getAuthHeader() }
        );
        await this.disconnectMqtt();
        return response.success;
    }

    async getActiveUsers(documentId: string): Promise<ActiveUser[]> {
        const response = await this.client.getActiveUsers(
            { documentId },
            { headers: getAuthHeader() }
        );
        return response.users;
    }

    async syncDocument(
        documentId: string,
        operations: Operation[],
        baseVersion: string
    ): Promise<SyncDocumentResponse> {
        return await this.client.syncDocument(
            { documentId, operations, baseVersion },
            { headers: getAuthHeader() }
        );
    }

    onDocumentChange(documentId: string, callback: (change: DocumentChange) => void) {
        this.subscriptions.set(documentId, callback);
    }

    private async connectMqtt(topic: string): Promise<void> {
        const clientId = `syncwrite_${Math.random().toString(16).substr(2, 8)}`;

        this.mqttClient = connect({
            hostname: import.meta.env.VITE_MQTT_BROKER,
            port: Number(import.meta.env.VITE_MQTT_PORT),
            path: import.meta.env.VITE_MQTT_PATH,
            clientId,
            clean: true,
            protocol: 'wss',
        });

        return new Promise((resolve, reject) => {
            this.mqttClient.on('connect', () => {
                this.mqttClient.subscribe(topic, (err: Error) => {
                    if (err) {
                        reject(err);
                    } else {
                        resolve();
                    }
                });
            });

            this.mqttClient.on('message', (_topic: string, message: Buffer) => {
                const change = JSON.parse(message.toString()) as DocumentChange;
                const callback = this.subscriptions.get(change.documentId);
                if (callback) {
                    callback(change);
                }
            });

            this.mqttClient.on('error', (err: Error) => {
                reject(err);
            });
        });
    }

    private async disconnectMqtt(): Promise<void> {
        if (this.mqttClient) {
            return new Promise((resolve) => {
                this.mqttClient.end(false, {}, () => {
                    this.mqttClient = null;
                    this.subscriptions.clear();
                    resolve();
                });
            });
        }
    }
}

export const collaborationService = new CollaborationService();