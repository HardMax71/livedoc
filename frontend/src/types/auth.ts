export interface User {
    id: string;
    email: string;
    username: string;
    createdAt: Date;
    updatedAt: Date;
}

export interface LoginRequest {
    email: string;
    password: string;
}

export interface LoginResponse {
    user: User;
    accessToken: string;
    refreshToken: string;
}

export interface RegisterRequest {
    email: string;
    username: string;
    password: string;
}

export interface RegisterResponse {
    user: User;
    accessToken: string;
    refreshToken: string;
}

export interface RefreshRequest {
    refreshToken: string;
}

export interface RefreshResponse {
    accessToken: string;
    refreshToken: string;
}