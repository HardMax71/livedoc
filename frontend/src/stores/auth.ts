import { createStore } from 'solid-js/store';
import { User } from '../types/auth';
import { authService } from '../api/auth';

interface AuthState {
    user: User | null;
    isAuthenticated: boolean;
    isLoading: boolean;
    error: string | null;
}

const initialState: AuthState = {
    user: null,
    isAuthenticated: false,
    isLoading: false,
    error: null,
};

const [state, setState] = createStore(initialState);

export const authStore = {
    get state() {
        return state;
    },

    async login(email: string, password: string) {
        setState({ isLoading: true, error: null });
        try {
            const response = await authService.login({ email, password });
            setState({
                user: response.user,
                isAuthenticated: true,
                isLoading: false,
            });
        } catch (error) {
            setState({
                error: error instanceof Error ? error.message : 'Login failed',
                isLoading: false,
            });
        }
    },

    async register(email: string, username: string, password: string) {
        setState({ isLoading: true, error: null });
        try {
            const response = await authService.register({ email, username, password });
            setState({
                user: response.user,
                isAuthenticated: true,
                isLoading: false,
            });
        } catch (error) {
            setState({
                error: error instanceof Error ? error.message : 'Registration failed',
                isLoading: false,
            });
        }
    },

    async logout() {
        setState({ isLoading: true, error: null });
        try {
            await authService.logout();
            setState({
                user: null,
                isAuthenticated: false,
                isLoading: false,
            });
        } catch (error) {
            setState({
                error: error instanceof Error ? error.message : 'Logout failed',
                isLoading: false,
            });
        }
    },

    async refreshToken() {
        const refreshToken = localStorage.getItem('refreshToken');
        if (!refreshToken) return;

        try {
            await authService.refresh({ refreshToken });
        } catch (error) {
            setState({
                user: null,
                isAuthenticated: false,
                error: 'Session expired',
            });
        }
    },

    clearError() {
        setState({ error: null });
    },
};