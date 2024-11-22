import { createClient, getAuthHeader } from './grpc-client';
import type {
    LoginRequest,
    LoginResponse,
    RegisterRequest,
    RegisterResponse,
    RefreshRequest,
    RefreshResponse
} from '../types/auth';

class AuthService {
    private client = createClient(AuthService);

    async login(request: LoginRequest): Promise<LoginResponse> {
        const response = await this.client.login(request);
        this.saveTokens(response.accessToken, response.refreshToken);
        return response;
    }

    async register(request: RegisterRequest): Promise<RegisterResponse> {
        const response = await this.client.register(request);
        this.saveTokens(response.accessToken, response.refreshToken);
        return response;
    }

    async refresh(request: RefreshRequest): Promise<RefreshResponse> {
        const response = await this.client.refresh(request);
        this.saveTokens(response.accessToken, response.refreshToken);
        return response;
    }

    async logout(): Promise<void> {
        const refreshToken = localStorage.getItem('refreshToken');
        if (refreshToken) {
            await this.client.logout({ refreshToken }, { headers: getAuthHeader() });
        }
        this.clearTokens();
    }

    private saveTokens(accessToken: string, refreshToken: string): void {
        localStorage.setItem('accessToken', accessToken);
        localStorage.setItem('refreshToken', refreshToken);
    }

    private clearTokens(): void {
        localStorage.removeItem('accessToken');
        localStorage.removeItem('refreshToken');
    }
}

export const authService = new AuthService();