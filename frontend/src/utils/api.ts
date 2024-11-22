import { authStore } from '../stores/auth';

export const API_URL = import.meta.env.VITE_API_URL;

interface FetchOptions extends RequestInit {
    skipAuth?: boolean;
}

export async function fetchApi<T>(
    endpoint: string,
    options: FetchOptions = {}
): Promise<T> {
    const { skipAuth = false, ...fetchOptions } = options;

    const headers = new Headers(fetchOptions.headers);

    if (!skipAuth) {
        const token = localStorage.getItem('accessToken');
        if (token) {
            headers.set('Authorization', `Bearer ${token}`);
        }
    }

    try {
        const response = await fetch(`${API_URL}${endpoint}`, {
            ...fetchOptions,
            headers,
        });

        if (response.status === 401) {
            // Try to refresh the token
            const refreshToken = localStorage.getItem('refreshToken');
            if (refreshToken) {
                await authStore.refreshToken();
                // Retry the original request
                return fetchApi(endpoint, options);
            } else {
                authStore.logout();
                throw new Error('Authentication required');
            }
        }

        if (!response.ok) {
            const error = await response.json().catch(() => ({}));
            throw new Error(error.message || 'An error occurred');
        }

        return response.json();
    } catch (error) {
        if (error instanceof Error) {
            throw error;
        }
        throw new Error('Network error');
    }
}