import { createConnectTransport } from '@bufbuild/connect-web';
import { createPromiseClient, PromiseClient } from '@bufbuild/connect';
import { ServiceType } from '@bufbuild/protobuf';

const transport = createConnectTransport({
    baseUrl: import.meta.env.VITE_API_URL,
    credentials: 'include',
});

export function createClient<T extends ServiceType>(
    service: T
): PromiseClient<T> {
    return createPromiseClient(service, transport);
}

export function getAuthHeader() {
    const token = localStorage.getItem('accessToken');
    return token ? { Authorization: `Bearer ${token}` } : {};
}