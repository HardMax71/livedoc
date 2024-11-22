import { createStore } from 'solid-js/store';
import {
    ActiveUser,
    DocumentChange,
    Operation
} from '../types/collaboration';
import { collaborationService } from '../api/collaboration';

interface CollaborationState {
    activeUsers: ActiveUser[];
    sessionId: string | null;
    mqttTopic: string | null;
    isConnected: boolean;
    isLoading: boolean;
    error: string | null;
}

const initialState: CollaborationState = {
    activeUsers: [],
    sessionId: null,
    mqttTopic: null,
    isConnected: false,
    isLoading: false,
    error: null,
};

const [state, setState] = createStore(initialState);

export const collaborationStore = {
    get state() {
        return state;
    },

    async joinSession(documentId: string) {
        setState({ isLoading: true, error: null });
        try {
            const response = await collaborationService.joinSession(documentId);
            setState({
                sessionId: response.sessionId,
                activeUsers: response.activeUsers,
                mqttTopic: response.mqttTopic,
                isConnected: true,
                isLoading: false,
            });
        } catch (error) {
            setState({
                error: error instanceof Error ? error.message : 'Failed to join session',
                isLoading: false,
            });
        }
    },

    async leaveSession() {
        if (!state.sessionId || !state.mqttTopic) return;

        setState({ isLoading: true, error: null });
        try {
            await collaborationService.leaveSession(state.sessionId, state.mqttTopic);
            setState({
                sessionId: null,
                mqttTopic: null,
                activeUsers: [],
                isConnected: false,
                isLoading: false,
            });
        } catch (error) {
            setState({
                error: error instanceof Error ? error.message : 'Failed to leave session',
                isLoading: false,
            });
        }
    },

    async syncDocument(documentId: string, operations: Operation[], baseVersion: string) {
        setState({ isLoading: true, error: null });
        try {
            const response = await collaborationService.syncDocument(
                documentId,
                operations,
                baseVersion
            );
            setState({ isLoading: false });
            return response;
        } catch (error) {
            setState({
                error: error instanceof Error ? error.message : 'Failed to sync document',
                isLoading: false,
            });
        }
    },

    onDocumentChange(documentId: string, callback: (change: DocumentChange) => void) {
        collaborationService.onDocumentChange(documentId, callback);
    },

    updateActiveUsers(users: ActiveUser[]) {
        setState({ activeUsers: users });
    },

    clearError() {
        setState({ error: null });
    },
};